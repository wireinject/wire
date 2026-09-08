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

type inner struct{ N int }
type S struct{ inner }

func provide() inner { return inner{7} }
func main() {
	if inject().N != 7 || injectField(S{inner{7}}).N != 7 {
		panic("private field")
	}
	fmt.Println("ok")
}
