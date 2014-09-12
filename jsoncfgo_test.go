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
*/

package jsoncfgo

import (
	"os"
	"reflect"
	"strings"
	"testing"
	u "github.com/go-goodies/go_utils"
)

// TESTS
func TestIncludes(t *testing.T) {
	obj, err := ReadFile("testdata/include1.json")
	if err != nil {
		t.Fatal(err)
	}
	two := obj.RequiredObject("two")
	if err := obj.Validate(); err != nil {
		t.Error(err)
	}
	if g, e := two.RequiredString("key"), "value"; g != e {
		t.Errorf("sub object key = %q; want %q", g, e)
	}
}

func TestIncludeLoop(t *testing.T) {
	_, err := ReadFile("testdata/loop1.json")
	if err == nil {
		t.Fatal("expected an error about import cycles.")
	}
	if !strings.Contains(err.Error(), "include cycle detected") {
		t.Fatalf("expected an error about import cycles; got: %v", err)
	}
}

func TestBoolEnvs(t *testing.T) {
	os.Setenv("TEST_EMPTY", "")
	os.Setenv("TEST_TRUE", "true")
	os.Setenv("TEST_ONE", "1")
	os.Setenv("TEST_ZERO", "0")
	os.Setenv("TEST_FALSE", "false")
	obj, err := ReadFile("testdata/boolenv.json")
	if err != nil {
		t.Fatal(err)
	}
	if str := obj.RequiredString("emptystr"); str != "" {
		t.Errorf("str = %q, want empty", str)
	}
	tests := []struct {
		key  string
		want bool
	}{
		{"def_false", false},
		{"def_true", true},
		{"set_true_def_false", true},
		{"set_false_def_true", false},
		{"lit_true", true},
		{"lit_false", false},
		{"one", true},
		{"zero", false},
	}
	for _, tt := range tests {
		if v := obj.RequiredBool(tt.key); v != tt.want {
			t.Errorf("key %q = %v; want %v", tt.key, v, tt.want)
		}
	}
	if err := obj.Validate(); err != nil {
		t.Error(err)
	}
}

func TestListExpansion(t *testing.T) {
	os.Setenv("TEST_BAR", "bar")
	obj, err := ReadFile("testdata/listexpand.json")
	if err != nil {
		t.Fatal(err)
	}
	s := obj.RequiredString("str")
	l := obj.RequiredList("list")
	if err := obj.Validate(); err != nil {
		t.Error(err)
	}
	want := []string{"foo", "bar"}
	if !reflect.DeepEqual(l, want) {
		t.Errorf("got = %#v\nwant = %#v", l, want)
	}
	if s != "bar" {
		t.Errorf("str = %q, want %q", s, "bar")
	}
}

func TestUint(t *testing.T) {
	obj, err := ReadFile("testdata/uint.json")
	if err != nil {
		t.Fatal(err)
	}
	myuint := obj.RequiredUint("myuint")
	myuint2 := obj.RequiredUint("myuint2")
	myint := obj.RequiredInt("myint")
	mystring := obj.RequiredString("mystring")
	if err := obj.Validate(); err != nil {
		t.Error(err)
	}
	if !u.IsUint(myuint) {
		t.Errorf("%v should have been of type uint", myuint)
	}
	if !u.IsUint(myuint2) {
		t.Errorf("%v should have been of type uint", myuint2)
	}
	if u.IsUint(myint) {
		t.Errorf("%v should not have been of type uint", myint)
	}
	if u.IsUint(mystring) {
		t.Errorf("%v should not have been of type uint", mystring)
	}
}

