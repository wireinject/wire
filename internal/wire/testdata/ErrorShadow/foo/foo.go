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

var err error
var failure = errors.New("provider failure")

func provide(fail bool) (int, error) {
	if fail {
		return 0, failure
	}
	return 7, nil
}
func main() {
	v, e := inject(true)
	if v != 0 || e != failure {
		panic("provider error lost")
	}
	v, e = inject(false)
	if v != 7 || e != nil {
		panic("success changed")
	}
	fmt.Println("ok")
}
