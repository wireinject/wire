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

package other

import "github.com/google/wire"

func private() int { return 7 }

type hidden struct{ V int }

func Make[T any]() int { return 7 }

var FuncSet = wire.NewSet(private)
var StructSet = wire.NewSet(wire.Value(7), wire.Struct(new(hidden), "*"), wire.Bind(new(any), new(hidden)))
var ArgumentSet = wire.NewSet(Make[map[string][]*hidden])

var AnonymousFieldSet = wire.NewSet(Make[struct{ private int }])
var AnonymousMethodSet = wire.NewSet(Make[interface{ private() }])
