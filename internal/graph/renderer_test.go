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
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestBuiltinFormats(t *testing.T) {
	want := []string{"dot", "mermaid"}
	got := Formats()
	if len(got) < len(want) {
		t.Fatalf("Formats() = %v, want it to contain %v", got, want)
	}
	for _, format := range want {
		if _, ok := Lookup(format); !ok {
			t.Errorf("Lookup(%q) = false, want true", format)
		}
	}
	if _, ok := Lookup("MerMaid"); !ok {
		t.Error(`Lookup("MerMaid") = false, want true; lookup should be case-insensitive`)
	}
	if _, ok := Lookup("nope"); ok {
		t.Error(`Lookup("nope") = true, want false`)
	}
}

func TestRenderUnknownFormat(t *testing.T) {
	err := Render(io.Discard, "nope", nil)
	if err == nil {
		t.Fatal("Render with an unknown format returned nil error")
	}
	if !strings.Contains(err.Error(), "mermaid") {
		t.Errorf("error %q does not list the available formats", err)
	}
}

// listRenderer is a minimal custom format, standing in for the renderers a
// future change might add.
type listRenderer struct{}

func (listRenderer) Format() string { return "testlist" }

func (listRenderer) Render(w io.Writer, graphs []*Graph) error {
	for _, g := range graphs {
		for _, n := range g.Nodes {
			if _, err := fmt.Fprintf(w, "%s: %s\n", g.Target, n.Label); err != nil {
				return err
			}
		}
	}
	return nil
}

func TestRegisterCustomRenderer(t *testing.T) {
	Register(listRenderer{})
	t.Cleanup(func() { delete(renderers, "testlist") })

	var sb strings.Builder
	g := &Graph{Target: "buildApp", Nodes: []*Node{{ID: "n0", Label: "*foo.App"}}}
	if err := Render(&sb, "testlist", []*Graph{g}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if got, want := sb.String(), "buildApp: *foo.App\n"; got != want {
		t.Errorf("Render wrote %q, want %q", got, want)
	}
}
