// Copyright 2026 The Wire Authors
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

package graph

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestDotQuote(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"plain", `"plain"`},
		{`a"b`, `"a\"b"`},
		{`a\b`, `"a\\b"`},
		{"a\nb", `"a\nb"`},
	}
	for _, tt := range tests {
		if got := dotQuote(tt.in); got != tt.want {
			t.Errorf("dotQuote(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestDotRender(t *testing.T) {
	g := &Graph{
		Target: "buildApp",
		Nodes: []*Node{
			{ID: "n0", Label: "*foo.Config", Pkg: "pkg/foo", Kind: KindInput},
			{ID: "n1", Label: "*foo.App", Pkg: "pkg/foo", Kind: KindProvider},
		},
		Edges: []*Edge{{From: "n0", To: "n1", Label: "foo.New"}},
	}
	var sb strings.Builder
	if err := Render(&sb, "dot", []*Graph{g}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := sb.String()
	for _, want := range []string{`digraph "buildApp" {`, `"n0" [label="*foo.Config", shape=ellipse];`, `"n0" -> "n1" [label="foo.New"];`} {
		if !strings.Contains(out, want) {
			t.Errorf("dot output missing %q:\n%s", want, out)
		}
	}
}

func TestDotRenderRejectsMultipleGraphs(t *testing.T) {
	graphs := []*Graph{{Target: "buildA"}, {Target: "buildB"}}
	err := Render(io.Discard, "dot", graphs)
	if !errors.Is(err, ErrMultipleGraphs) {
		t.Errorf("Render with 2 graphs returned %v, want an error wrapping ErrMultipleGraphs", err)
	}
}
