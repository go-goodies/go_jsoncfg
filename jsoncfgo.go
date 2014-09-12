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
* Added convenience functions:
*  Bool, Int, Int64, IntList, List, Load, Object, String, requiredIntList
*/

// Package jsoncfgo defines a helper type for JSON objects to be
// used for configuration.
package jsoncfgo

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"log"
)

// Load calls ReadFile to read json config data from the file specified by configPath
// If an error occurs, it will log the error.
func Load(configPath string) Obj {
	conf, err := ReadFile(configPath)
	if err != nil {
		log.Fatal(err.Error())
	}
	return Obj(conf)
}

// Obj is a JSON configuration map.
type Obj map[string]interface{}

// Reads json config data from the specified open file, expanding
// all expressions
func ReadFile(configPath string) (Obj, error) {
	var c ConfigParser
	return c.ReadFile(configPath)
}

func (jc Obj) RequiredObject(key string) Obj {
	return jc.obj(key, false)
}

func (jc Obj) OptionalObject(key string) Obj {
	return jc.obj(key, true)
}

// Object is an OptionalObject
func (jc Obj) Object(key string) Obj {
	return jc.obj(key, true)
}

func (jc Obj) obj(key string, optional bool) Obj {
	jc.noteKnownKey(key)
	ei, ok := jc[key]
	if !ok {
		if optional {
			return make(Obj)
		}
		jc.appendError(fmt.Errorf("Missing required config key %q (object)", key))
		return make(Obj)
	}
	m, ok := ei.(map[string]interface{})
	if !ok {
		jc.appendError(fmt.Errorf("Expected config key %q to be an object, not %T", key, ei))
		return make(Obj)
	}
	return Obj(m)
}

func (jc Obj) RequiredString(key string) string {
	return jc.string(key, nil)
}

func (jc Obj) OptionalString(key, def string) string {
	return jc.string(key, &def)
}

// String is an OptionalString and accepts an optional string parameter
func (jc Obj) String(key string, args ...interface{}) string {
	def := ""
	for _, arg := range args {
		switch t := arg.(type) {
		case string:
			def = t
		default:
			panic(fmt.Sprintf("ERROR - Invalid argument (%v).  Must be a string.", arg))
		}
	}
	return jc.string(key, &def)
}

func (jc Obj) string(key string, def *string) string {
	jc.noteKnownKey(key)
	ei, ok := jc[key]
	if !ok {
		if def != nil {
			return *def
		}
		jc.appendError(fmt.Errorf("Missing required config key %q (string)", key))
		return ""
	}
	s, ok := ei.(string)
	if !ok {
		jc.appendError(fmt.Errorf("Expected config key %q to be a string", key))
		return ""
	}
	return s
}

func (jc Obj) RequiredStringOrObject(key string) interface{} {
	return jc.stringOrObject(key, true)
}

func (jc Obj) OptionalStringOrObject(key string) interface{} {
	return jc.stringOrObject(key, false)
}

// StringOrObject is an OptionalStringOrObject
func (jc Obj) StringOrObject(key string) interface{} {
	return jc.stringOrObject(key, false)
}

func (jc Obj) stringOrObject(key string, required bool) interface{} {
	jc.noteKnownKey(key)
	ei, ok := jc[key]
	if !ok {
		if !required {
			return nil
		}
		jc.appendError(fmt.Errorf("Missing required config key %q (string or object)", key))
		return ""
	}
	if _, ok := ei.(map[string]interface{}); ok {
		return ei
	}
	if _, ok := ei.(string); ok {
		return ei
	}
	jc.appendError(fmt.Errorf("Expected config key %q to be a string or object", key))
	return ""
}

func (jc Obj) RequiredBool(key string) bool {
	return jc.bool(key, nil)
}

func (jc Obj) OptionalBool(key string, def bool) bool {
	return jc.bool(key, &def)
}

// Bool is an OptionalBool
func (jc Obj) Bool(key string, args ...interface{}) bool {
	def := false
	for _, arg := range args {
		switch t := arg.(type) {
		case bool:
			def = t
		default:
			panic(fmt.Sprintf("ERROR - Invalid argument (%v).  Must be a bool.", arg))
		}
	}
	return jc.bool(key, &def)
}

func (jc Obj) bool(key string, def *bool) bool {
	jc.noteKnownKey(key)
	ei, ok := jc[key]
	if !ok {
		if def != nil {
			return *def
		}
		jc.appendError(fmt.Errorf("Missing required config key %q (boolean)", key))
		return false
	}
	switch v := ei.(type) {
	case bool:
		return v
	case string:
		b, err := strconv.ParseBool(v)
		if err != nil {
			jc.appendError(fmt.Errorf("Config key %q has bad boolean format %q", key, v))
		}
		return b
	default:
		jc.appendError(fmt.Errorf("Expected config key %q to be a boolean", key))
		return false
	}
}

func (jc Obj) RequiredInt(key string) int {
	return jc.int(key, nil)
}

func (jc Obj) OptionalInt(key string, def int) int {
	return jc.int(key, &def)
}

// Int is an OptionalInt
func (jc Obj) Int(key string, args ...interface{}) int {
	def := 0
	for _, arg := range args {
		switch t := arg.(type) {
		case int:
			def = t
		default:
			panic(fmt.Sprintf("ERROR - Invalid argument (%v).  Must be an int.", arg))
		}
	}
	return jc.int(key, &def)
}

func (jc Obj) int(key string, def *int) int {
	jc.noteKnownKey(key)
	ei, ok := jc[key]
	if !ok {
		if def != nil {
			return *def
		}
		jc.appendError(fmt.Errorf("Missing required config key %q (integer)", key))
		return 0
	}
	b, ok := ei.(float64)
	if !ok {
		jc.appendError(fmt.Errorf("Expected config key %q to be a number", key))
		return 0
	}
	return int(b)
}


func (jc Obj) RequiredUint(key string) uint {
	return jc.uint(key, nil)
}

func (jc Obj) OptionalUint(key string, def uint) uint {
	return jc.uint(key, &def)
}

// Uint is an Optional Uint
func (jc Obj) Uint(key string, args ...interface{}) uint {
	var def uint
	for _, arg := range args {
		switch t := arg.(type) {
		case int:
			def = uint(t)
		case uint:
			def = t
		default:
			panic(fmt.Sprintf("ERROR - Invalid argument (%v).  Must be an uint.", arg))
		}
	}
	return jc.uint(key, &def)
}

func (jc Obj) uint(key string, def *uint) uint {
	jc.noteKnownKey(key)
	ei, ok := jc[key]
	if !ok {
		if def != nil {
			return *def
		}
		jc.appendError(fmt.Errorf("Missing required config key %q (integer)", key))
		return 0
	}
	b, ok := ei.(float64)
	if !ok {
		jc.appendError(fmt.Errorf("Expected config key %q to be a number", key))
		return 0
	}
	return uint(b)
}



func (jc Obj) RequiredInt64(key string) int64 {
	return jc.int64(key, nil)
}

func (jc Obj) OptionalInt64(key string, def int64) int64 {
	return jc.int64(key, &def)
}

// Int64 is an Optional Int64
func (jc Obj) Int64(key string, args ...interface{}) int64 {
	var def int64
	for _, arg := range args {
		switch t := arg.(type) {
		case int:
			def = int64(t)
		case int64:
			def = t
		default:
			panic(fmt.Sprintf("ERROR - Invalid argument (%v).  Must be an int64.", arg))
		}
	}
	return jc.int64(key, &def)
}

func (jc Obj) int64(key string, def *int64) int64 {
	jc.noteKnownKey(key)
	ei, ok := jc[key]
	if !ok {
		if def != nil {
			return *def
		}
		jc.appendError(fmt.Errorf("Missing required config key %q (integer)", key))
		return 0
	}
	b, ok := ei.(float64)
	if !ok {
		jc.appendError(fmt.Errorf("Expected config key %q to be a number", key))
		return 0
	}
	return int64(b)
}


func (jc Obj) RequiredList(key string) []string {
	return jc.requiredList(key, true)
}

// List accepts and optional parameter of type []string
func (jc Obj) List(key string, args ...interface{}) []string {
	ret := jc.requiredList(key, true)
	if ret == nil {
		ret = args[0].([]string)
	}
	return ret
}

// List accepts and optional parameter of type []int64
func (jc Obj) IntList(key string, args ...interface{}) []int64 {
	ret := jc.requiredIntList(key, true)
	if ret == nil {
		ret = args[0].([]int64)
	}
	return ret
}

func (jc Obj) OptionalList(key string) []string {
	return jc.requiredList(key, false)
}

func (jc Obj) requiredList(key string, required bool) []string {
	jc.noteKnownKey(key)
	ei, ok := jc[key]
	if !ok {
		if required {
			jc.appendError(fmt.Errorf("Missing required config key %q (list of strings)", key))
		}
		return nil
	}
	eil, ok := ei.([]interface{})
	if !ok {
		jc.appendError(fmt.Errorf("Expected config key %q to be a list, not %T", key, ei))
		return nil
	}
	sl := make([]string, len(eil))
	for i, ei := range eil {
		s, ok := ei.(string)
		if !ok {
			jc.appendError(fmt.Errorf("Expected config key %q index %d to be a string, not %T", key, i, ei))
			return nil
		}
		sl[i] = s
	}
	return sl
}

func (jc Obj) requiredIntList(key string, required bool) []int64 {
	jc.noteKnownKey(key)
	ei, ok := jc[key]
	if !ok {
		if required {
			jc.appendError(fmt.Errorf("Missing required config key %q (list of ints)", key))
		}
		return nil
	}
	eil, ok := ei.([]interface{})
	if !ok {
		jc.appendError(fmt.Errorf("Expected config key %q to be a list, not %T", key, ei))
		return nil
	}
	sl := make([]int64, len(eil))
	for i, ei := range eil {
		sl[i] = int64(ei.(float64))
	}
	return sl
}

func (jc Obj) noteKnownKey(key string) {
	_, ok := jc["_knownkeys"]
	if !ok {
		jc["_knownkeys"] = make(map[string]bool)
	}
	jc["_knownkeys"].(map[string]bool)[key] = true
}

func (jc Obj) appendError(err error) {
	ei, ok := jc["_errors"]
	if ok {
		jc["_errors"] = append(ei.([]error), err)
	} else {
		jc["_errors"] = []error{err}
	}
}

// UnknownKeys returns the keys from the config that have not yet been discovered by one of the RequiredT or OptionalT calls.
func (jc Obj) UnknownKeys() []string {
	ei, ok := jc["_knownkeys"]
	var known map[string]bool
	if ok {
		known = ei.(map[string]bool)
	}
	var unknown []string
	for k, _ := range jc {
		if ok && known[k] {
			continue
		}
		if strings.HasPrefix(k, "_") {
			// Permit keys with a leading underscore as a
			// form of comments.
			continue
		}
		unknown = append(unknown, k)
	}
	sort.Strings(unknown)
	return unknown
}

func (jc Obj) Validate() error {
	unknown := jc.UnknownKeys()
	for _, k := range unknown {
		jc.appendError(fmt.Errorf("Unknown key %q", k))
	}

	ei, ok := jc["_errors"]
	if !ok {
		return nil
	}
	errList := ei.([]error)
	if len(errList) == 1 {
		return errList[0]
	}
	strs := make([]string, 0)
	for _, v := range errList {
		strs = append(strs, v.Error())
	}
	return fmt.Errorf("Multiple errors: " + strings.Join(strs, ", "))
}
