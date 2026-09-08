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

package main

import "github.com/google/wire"

func injectBox() *Box[int] { wire.Build(wire.Value(7), wire.Struct(new(Box[int]), "*")); return nil }
func injectAlias() Box[int] {
	wire.Build(wire.Value(7), wire.Struct(new(IntBox), "*"))
	return IntBox{}
}
func injectPair() Pair[int, string] {
	wire.Build(wire.Value(7), wire.Value("ok"), wire.Struct(new(Pair[int, string]), "*"))
	return Pair[int, string]{}
}
