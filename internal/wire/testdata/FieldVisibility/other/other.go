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

type S struct{ private int }

func New() S { return S{7} }

type inner struct{ N int }
type Embedded struct{ inner }

func NewInner() inner { return inner{7} }

var StructSet = wire.NewSet(wire.Struct(new(S), "*"))
var FieldSet = wire.NewSet(New, wire.FieldsOf(new(S), "private"))
