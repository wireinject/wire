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

//+build wireinject

// All declarations are in one file so that Wire must copy the non-injector
// generic declarations into wire_gen.go.

package main

import (
	"fmt"

	"github.com/google/wire"
)

type Box[T any] struct {
	V T
}

type Pair[A, B any] struct {
	A A
	B B
}

// defaultPair and Identity are non-injector declarations that exercise
// generic type parameters and multi-argument instantiations when copied.
var defaultPair = Pair[int, string]{A: 1, B: "two"}

func Identity[T any](v T) T {
	return v
}

func MakeBox[T any](v T) Box[T] {
	return Box[T]{V: v}
}

func provideInt() int {
	return 42
}

func main() {
	box := injectBox()
	fmt.Println(box.V, defaultPair.A, defaultPair.B, Identity(7))
}

func injectBox() Box[int] {
	wire.Build(provideInt, MakeBox[int])
	return Box[int]{}
}
