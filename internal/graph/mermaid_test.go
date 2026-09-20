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
	"strings"
	"testing"
)

func TestEscapeMermaid(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"plain", "plain"},
		{"map[string]string", "map[string]string"},
		{`*foo.Bar`, `*foo.Bar`},
		{`a"b`, "a#quot;b"},
		{"a|b", "a#124;b"},
		{"a<b>&c", "a#60;b#62;#38;c"},
		{"a\nb", "a b"},
	}
	for _, tt := range tests {
		if got := escapeMermaid(tt.in); got != tt.want {
			t.Errorf("escapeMermaid(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestMermaidNode(t *testing.T) {
	tests := []struct {
		kind  Kind
		label string
		want  string
	}{
		{KindProvider, "*foo.Bar", `["*foo.Bar"]`},
		{KindInput, "*foo.Bar", `(["*foo.Bar"])`},
		{KindInterface, "foo.Iface", `[["foo.Iface"]]`},
		{KindValue, "string", `{{"string"}}`},
		{KindField, "int", `[/"int"/]`},
	}
	for _, tt := range tests {
		n := &Node{ID: "n0", Label: tt.label, Kind: tt.kind}
		if got := mermaidNode(n); got != tt.want {
			t.Errorf("mermaidNode(kind=%d, %q) = %q, want %q", tt.kind, tt.label, got, tt.want)
		}
	}
}

func TestMermaidRenderIsFenced(t *testing.T) {
	g := &Graph{
		Target: "buildApp",
		Nodes: []*Node{
			{ID: "n0", Label: "*foo.Config", Pkg: "pkg/foo", Kind: KindInput},
			{ID: "n1", Label: "*foo.App", Pkg: "pkg/foo", Kind: KindProvider},
		},
		Edges: []*Edge{{From: "n0", To: "n1", Label: "foo.New"}},
	}
	var sb strings.Builder
	if err := Render(&sb, "mermaid", []*Graph{g}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := sb.String()
	for _, want := range []string{"## buildApp", "```mermaid", "flowchart TD", "```\n", `n0 -->|foo.New| n1`} {
		if !strings.Contains(out, want) {
			t.Errorf("mermaid output missing %q:\n%s", want, out)
		}
	}
}

func TestMermaidRenderMultipleGraphs(t *testing.T) {
	graphs := []*Graph{
		{Target: "buildA", Nodes: []*Node{{ID: "n0", Label: "*foo.A", Pkg: "pkg/foo"}}},
		{Target: "buildB", Nodes: []*Node{{ID: "n0", Label: "*foo.B", Pkg: "pkg/foo"}}},
	}
	var sb strings.Builder
	if err := Render(&sb, "mermaid", graphs); err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := sb.String()
	if got := strings.Count(out, "```mermaid"); got != 2 {
		t.Errorf("got %d Mermaid diagrams, want 2:\n%s", got, out)
	}
	if !strings.Contains(out, "## buildA") || !strings.Contains(out, "## buildB") {
		t.Errorf("output missing target headings:\n%s", out)
	}
}
