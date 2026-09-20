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
	"testing"

	"github.com/wireinject/wire/internal/wire"
)

func TestGroups(t *testing.T) {
	g := &Graph{
		Nodes: []*Node{
			{ID: "n0", Label: "b.Type", Pkg: "pkg/b"},
			{ID: "n1", Label: "a.Type", Pkg: "pkg/a"},
			{ID: "n2", Label: "a.Type2", Pkg: "pkg/a"},
		},
	}
	groups := g.Groups()
	if len(groups) != 2 {
		t.Fatalf("Groups returned %d groups, want 2", len(groups))
	}
	if groups[0].Pkg != "pkg/a" || groups[1].Pkg != "pkg/b" {
		t.Fatalf("groups not sorted by package: %q, %q", groups[0].Pkg, groups[1].Pkg)
	}
	if got := []string{groups[0].Nodes[0].Label, groups[0].Nodes[1].Label}; got[0] != "a.Type" || got[1] != "a.Type2" {
		t.Errorf("nodes not sorted by label within group: %v", got)
	}
}

func TestTargetMatching(t *testing.T) {
	in := &wire.Injector{ImportPath: "example.com/foo/app", FuncName: "buildApp"}
	for _, pattern := range []string{"buildApp", "example.com/foo/app.buildApp", `"example.com/foo/app".buildApp`} {
		if !injectorMatches(in, pattern) {
			t.Errorf("injectorMatches(%q) = false, want true", pattern)
		}
	}
	if injectorMatches(in, "other") {
		t.Error("injectorMatches(\"other\") = true, want false")
	}

	id := wire.ProviderSetID{ImportPath: "example.com/foo/app", VarName: "ProviderSet"}
	for _, pattern := range []string{"ProviderSet", "example.com/foo/app.ProviderSet", `"example.com/foo/app".ProviderSet`} {
		if !providerSetMatches(id, pattern) {
			t.Errorf("providerSetMatches(%q) = false, want true", pattern)
		}
	}
	if providerSetMatches(id, "OtherSet") {
		t.Error(`providerSetMatches("OtherSet") = true, want false`)
	}
}
