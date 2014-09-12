/*
Copyright 2011 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Changes made:
* Changed package name from jsonconfig to jsoncfgo
* Added Load function in jsoncfgo.go
*/

package jsoncfgo

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"camlistore.org/pkg/errorutil"
	"camlistore.org/pkg/osutil"
)

type stringVector struct {
	v []string
}

func (v *stringVector) Push(s string) {
	v.v = append(v.v, s)
}

func (v *stringVector) Pop() {
	v.v = v.v[:len(v.v)-1]
}

func (v *stringVector) Last() string {
	return v.v[len(v.v)-1]
}

// A File is the type returned by ConfigParser.Open.
type File interface {
	io.ReadSeeker
	io.Closer
	Name() string
}

// ConfigParser specifies the environment for parsing a config file
// and evaluating expressions.
type ConfigParser struct {
	rootJSON Obj

	touchedFiles map[string]bool
	includeStack stringVector

	// Open optionally specifies an opener function.
	Open func(filename string) (File, error)
}

func (c *ConfigParser) open(filename string) (File, error) {
	if c.Open == nil {
		return os.Open(filename)
	}
	return c.Open(filename)
}

// Validates variable names for config _env expresssions
var envPattern = regexp.MustCompile(`\$\{[A-Za-z0-9_]+\}`)

func (c *ConfigParser) ReadFile(path string) (m map[string]interface{}, err error) {
	c.touchedFiles = make(map[string]bool)
	c.rootJSON, err = c.recursiveReadJSON(path)
	return c.rootJSON, err
}

// Decodes and evaluates a json config file, watching for include cycles.
func (c *ConfigParser) recursiveReadJSON(configPath string) (decodedObject map[string]interface{}, err error) {

	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to expand absolute path for %s", configPath)
	}
	if c.touchedFiles[absConfigPath] {
		return nil, fmt.Errorf("ConfigParser include cycle detected reading config: %v",
			absConfigPath)
	}
	c.touchedFiles[absConfigPath] = true

	c.includeStack.Push(absConfigPath)
	defer c.includeStack.Pop()

	var f File
	if f, err = c.open(configPath); err != nil {
		return nil, fmt.Errorf("Failed to open config: %v", err)
	}
	defer f.Close()

	decodedObject = make(map[string]interface{})
	dj := json.NewDecoder(f)
	if err = dj.Decode(&decodedObject); err != nil {
		extra := ""
		if serr, ok := err.(*json.SyntaxError); ok {
			if _, serr := f.Seek(0, os.SEEK_SET); serr != nil {
				log.Fatalf("seek error: %v", serr)
			}
			line, col, highlight := errorutil.HighlightBytePosition(f, serr.Offset)
			extra = fmt.Sprintf(":\nError at line %d, column %d (file offset %d):\n%s",
				line, col, serr.Offset, highlight)
		}
		return nil, fmt.Errorf("error parsing JSON object in config file %s%s\n%v",
			f.Name(), extra, err)
	}

	if err = c.evaluateExpressions(decodedObject, nil, false); err != nil {
		return nil, fmt.Errorf("error expanding JSON config expressions in %s:\n%v",
			f.Name(), err)
	}

	return decodedObject, nil
}

type expanderFunc func(c *ConfigParser, v []interface{}) (interface{}, error)

func namedExpander(name string) (expanderFunc, bool) {
	switch name {
	case "_env":
		return expanderFunc((*ConfigParser).expandEnv), true
	case "_fileobj":
		return expanderFunc((*ConfigParser).expandFile), true
	}
	return nil, false
}

func (c *ConfigParser) evalValue(v interface{}) (interface{}, error) {
	sl, ok := v.([]interface{})
	if !ok {
		return v, nil
	}
	if name, ok := sl[0].(string); ok {
		if expander, ok := namedExpander(name); ok {
			newval, err := expander(c, sl[1:])
			if err != nil {
				return nil, err
			}
			return newval, nil
		}
	}
	for i, oldval := range sl {
		newval, err := c.evalValue(oldval)
		if err != nil {
			return nil, err
		}
		sl[i] = newval
	}
	return v, nil
}

// CheckTypes parses m and returns an error if it encounters a type or value
// that is not supported by this package.
func (c *ConfigParser) CheckTypes(m map[string]interface{}) error {
	return c.evaluateExpressions(m, nil, true)
}

// evaluateExpressions parses recursively m, populating it with the values
// that are found, unless testOnly is true.
func (c *ConfigParser) evaluateExpressions(m map[string]interface{}, seenKeys []string, testOnly bool) error {
	for k, ei := range m {
		thisPath := append(seenKeys, k)
		switch subval := ei.(type) {
		case string:
			continue
		case bool:
			continue
		case float64:
			continue
		case []interface{}:
			if len(subval) == 0 {
				continue
			}
			evaled, err := c.evalValue(subval)
			if err != nil {
				return fmt.Errorf("%s: value error %v", strings.Join(thisPath, "."), err)
			}
			if !testOnly {
				m[k] = evaled
			}
		case map[string]interface{}:
			if err := c.evaluateExpressions(subval, thisPath, testOnly); err != nil {
				return err
			}
		default:
			return fmt.Errorf("%s: unhandled type %T", strings.Join(thisPath, "."), ei)
		}
	}
	return nil
}

// Permit either:
//    ["_env", "VARIABLE"] (required to be set)
// or ["_env", "VARIABLE", "default_value"]
func (c *ConfigParser) expandEnv(v []interface{}) (interface{}, error) {
	hasDefault := false
	def := ""
	if len(v) < 1 || len(v) > 2 {
		return "", fmt.Errorf("_env expansion expected 1 or 2 args, got %d", len(v))
	}
	s, ok := v[0].(string)
	if !ok {
		return "", fmt.Errorf("Expected a string after _env expansion; got %#v", v[0])
	}
	boolDefault, wantsBool := false, false
	if len(v) == 2 {
		hasDefault = true
		switch vdef := v[1].(type) {
		case string:
			def = vdef
		case bool:
			wantsBool = true
			boolDefault = vdef
		default:
			return "", fmt.Errorf("Expected default value in %q _env expansion; got %#v", s, v[1])
		}
	}
	var err error
	expanded := envPattern.ReplaceAllStringFunc(s, func(match string) string {
		envVar := match[2 : len(match)-1]
		val := os.Getenv(envVar)
		// Special case:
		if val == "" && envVar == "USER" && runtime.GOOS == "windows" {
			val = os.Getenv("USERNAME")
		}
		if val == "" {
			if hasDefault {
				return def
			}
			err = fmt.Errorf("couldn't expand environment variable %q", envVar)
		}
		return val
	})
	if wantsBool {
		if expanded == "" {
			return boolDefault, nil
		}
		return strconv.ParseBool(expanded)
	}
	return expanded, err
}

func (c *ConfigParser) expandFile(v []interface{}) (exp interface{}, err error) {
	if len(v) != 1 {
		return "", fmt.Errorf("_file expansion expected 1 arg, got %d", len(v))
	}
	var incPath string
	if incPath, err = osutil.FindCamliInclude(v[0].(string)); err != nil {
		return "", fmt.Errorf("Included config does not exist: %v", v[0])
	}
	if exp, err = c.recursiveReadJSON(incPath); err != nil {
		return "", fmt.Errorf("In file included from %s:\n%v",
			c.includeStack.Last(), err)
	}
	return exp, nil
}
