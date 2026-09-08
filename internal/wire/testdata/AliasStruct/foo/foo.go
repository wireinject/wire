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

//go:build wireinject
// +build wireinject

package main

import (
	"fmt"

	"github.com/google/wire"
)

type S struct {
	N int
}

// SAlias and IntBox are type aliases; wire.Struct must accept them.
type SAlias = S

type Box[T any] struct {
	V T
}

type IntBox = Box[int]

type Anonymous = struct{ N int }

func provideInt() int {
	return 7
}

func main() {
	fmt.Println(injectAlias().N, injectIntBox().V, injectAnonymous().N)
}

// Return underlying types to keep the golden independent of go/types alias
// representation. The Struct markers must still resolve the aliases.
func injectAlias() S {
	wire.Build(provideInt, wire.Struct(new(SAlias), "*"))
	return SAlias{}
}

func injectIntBox() Box[int] {
	wire.Build(provideInt, wire.Struct(new(IntBox), "*"))
	return IntBox{}
}

func injectAnonymous() struct{ N int } {
	wire.Build(provideInt, wire.Struct(new(Anonymous), "*"))
	return Anonymous{}
}
