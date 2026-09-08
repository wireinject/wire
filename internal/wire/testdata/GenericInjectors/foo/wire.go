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
	ext "example.com/other"
	"github.com/google/wire"
)

func injectFirst() ext.Box[int]                      { wire.Build(ext.Make[int]); return ext.Box[int]{} }
func injectAny[T any]() (T, error)                   { wire.Build(fail[T]); return fail[T]() }
func injectInt[T ~int]() (T, error)                  { wire.Build(fail[T]); return fail[T]() }
func injectStruct[T ~struct{ N int }]() (T, error)   { wire.Build(fail[T]); return fail[T]() }
func injectErr[err any]() (*err, error)              { wire.Build(failPtr[err]); return nil, nil }
func injectZero[zero any]() (zero, error)            { wire.Build(fail[zero]); return fail[zero]() }
func injectNew[T any]() (T, error)                   { wire.Build(fail[T]); return fail[T]() }
func injectCleanup[cleanup any]() (*cleanup, func()) { wire.Build(resource[cleanup]); return nil, nil }
func injectV[v any]() *v                             { wire.Build(value[v]); return nil }
func injectUnused[T any]() int                       { wire.Build(wire.Value(7)); return 0 }
func injectConstraint[T ext.Number](v T) T           { wire.Build(); return v }
func injectImport[other any]() ext.Box[other]        { wire.Build(ext.Make[other]); return ext.Box[other]{} }
