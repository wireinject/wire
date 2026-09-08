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

package main

import (
	"fmt"

	"github.com/google/wire"
)

// x collides with the package-level x in the generated file's scope.
var x = 0

func helper() int {
	x2 := 1
	x := 2
	_ = x
	return x2
}

func provideInt() int {
	return helper()
}

func main() {
	fmt.Println(injectInt(), x)
}

func injectInt() int {
	wire.Build(provideInt)
	return 0
}
