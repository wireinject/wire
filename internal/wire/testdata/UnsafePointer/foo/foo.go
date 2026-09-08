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
	"unsafe"
)

var failure = errors.New("failure")

func provide() (unsafe.Pointer, error) { return nil, failure }
func main() {
	p, e := inject()
	if p != nil || e != failure {
		panic("unsafe pointer error")
	}
	fmt.Println("ok")
}
