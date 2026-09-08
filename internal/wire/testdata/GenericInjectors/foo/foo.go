// Copyright 2025 The Wire Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"fmt"
)

var failure = errors.New("failure")
var cleaned bool

// Shadowing the builtin must not break generation of a generic zero value.
var new int

func fail[T any]() (T, error)       { var z T; return z, failure }
func failPtr[T any]() (*T, error)   { return nil, failure }
func resource[T any]() (*T, func()) { var z T; return &z, func() { cleaned = true } }
func value[T any]() *T              { var z T; return &z }
func main() {
	if v, e := injectAny[int](); v != 0 || e != failure {
		panic("int zero")
	}
	if v, e := injectAny[string](); v != "" || e != failure {
		panic("string zero")
	}
	if v, e := injectInt[int](); v != 0 || e != failure {
		panic("constraint zero")
	}
	if v, e := injectStruct[struct{ N int }](); v.N != 0 || e != failure {
		panic("struct zero")
	}
	if v, e := injectErr[int](); v != nil || e != failure {
		panic("err collision")
	}
	if v, e := injectZero[int](); v != 0 || e != failure {
		panic("zero collision")
	}
	if v, e := injectNew[int](); v != 0 || e != failure {
		panic("new collision")
	}
	if v, c := injectCleanup[int](); v == nil || c == nil {
		panic("cleanup collision")
	} else {
		c()
	}
	if !cleaned {
		panic("cleanup not called")
	}
	if injectV[int]() == nil {
		panic("local collision")
	}
	if injectUnused[string]() != 7 {
		panic("unused type parameter")
	}
	if injectConstraint(7) != 7 {
		panic("constraint")
	}
	if injectFirst().V != 0 || injectImport[int]().V != 0 {
		panic("import collision")
	}
	fmt.Println("ok")
}
