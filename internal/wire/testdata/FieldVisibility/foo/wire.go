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

import (
	"example.com/other"
	"github.com/google/wire"
)

func injectStar() other.S {
	wire.Build(wire.Value(7), wire.Struct(new(other.S), "*"))
	return other.S{}
}

// A local alias must not hide the foreign field's declaring package.
func injectAlias() other.S { wire.Build(wire.Value(7), wire.Struct(new(Alias), "*")); return Alias{} }
func injectField() int     { wire.Build(other.New, wire.FieldsOf(new(other.S), "private")); return 0 }
func injectSet() other.S   { wire.Build(wire.Value(7), other.StructSet); return other.S{} }
func injectFieldSet() int  { wire.Build(other.FieldSet); return 0 }
func injectEmbedded() other.Embedded {
	wire.Build(other.NewInner, wire.Struct(new(other.Embedded), "*"))
	return other.Embedded{}
}
