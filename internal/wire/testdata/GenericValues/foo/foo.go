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

import "fmt"

type Box[T any] struct{ V T }
type Pair[A, B any] struct {
	A A
	B B
}

func main() {
	if injectBox().V != 7 {
		panic("box")
	}
	if p := injectPair(); p.A != 7 || p.B != "ok" {
		panic("pair")
	}
	fmt.Println("ok")
}
