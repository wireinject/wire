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

// Package graph builds dependency graphs from Wire injectors and provider
// sets and renders them in diagram formats such as Mermaid and Graphviz DOT.
//
// A graph is built in two steps: Targets selects the injectors or provider
// sets to graph, and Build walks a target into a format-independent Graph.
// Rendering is done by a Renderer, so new output formats can be added without
// touching the graph model; see renderer.go.
package graph

import (
	"fmt"
	"go/types"
	"sort"
	"strings"

	"github.com/wireinject/wire/internal/wire"
	"golang.org/x/tools/go/types/typeutil"
)

// Kind determines how a node is rendered.
type Kind int

const (
	KindProvider  Kind = iota // a type produced by a provider
	KindInterface             // a type produced via wire.Bind
	KindInput                 // an injector argument or an unbound input
	KindValue                 // a type produced by wire.Value or wire.InterfaceValue
	KindField                 // a type produced by wire.FieldsOf
)

// A Node is a single type in a dependency graph.
type Node struct {
	ID    string
	Label string
	Pkg   string // grouping key, an import path or a synthetic name
	Kind  Kind
}

// An Edge connects the node of a dependency to the node of the type that
// depends on it. Label names the provider or field that forms the connection.
type Edge struct {
	From, To string
	Label    string
}

// A Graph is a rendered-format-independent dependency graph.
type Graph struct {
	Target string
	Nodes  []*Node
	Edges  []*Edge
}

// A Group is a set of nodes that belong to the same package.
type Group struct {
	Pkg   string
	Nodes []*Node
}

// Groups groups the graph's nodes by package, sorting both the groups and the
// nodes within them so that output is deterministic.
func (g *Graph) Groups() []Group {
	byPkg := make(map[string][]*Node)
	for _, n := range g.Nodes {
		byPkg[n.Pkg] = append(byPkg[n.Pkg], n)
	}
	groups := make([]Group, 0, len(byPkg))
	for pkg, ns := range byPkg {
		sort.Slice(ns, func(i, j int) bool { return ns[i].Label < ns[j].Label })
		groups = append(groups, Group{Pkg: pkg, Nodes: ns})
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Pkg < groups[j].Pkg })
	return groups
}

// A Target is a single graph root: either an injector function or a top-level
// provider set.
type Target struct {
	Name  string
	Set   *wire.ProviderSet
	Roots []types.Type
}

// Options selects which injectors or provider sets to graph.
type Options struct {
	// Injector, if not empty, limits the targets to the injector with this
	// name, given as an unqualified or a qualified name.
	Injector string
	// Set, if not empty, graphs the top-level provider set with this variable
	// name or qualified name instead of graphing injectors.
	Set string
}

// Targets selects the injectors or provider sets to graph.
func Targets(info *wire.Info, opts Options) ([]Target, error) {
	if opts.Set != "" {
		var ids []wire.ProviderSetID
		for id := range info.Sets {
			if providerSetMatches(id, opts.Set) {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return nil, fmt.Errorf("no provider set matching %q", opts.Set)
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })
		targets := make([]Target, 0, len(ids))
		for _, id := range ids {
			set := info.Sets[id]
			targets = append(targets, Target{Name: id.String(), Set: set, Roots: set.Outputs()})
		}
		return targets, nil
	}

	injectors := append([]*wire.Injector(nil), info.Injectors...)
	sort.Slice(injectors, func(i, j int) bool {
		if injectors[i].ImportPath == injectors[j].ImportPath {
			return injectors[i].FuncName < injectors[j].FuncName
		}
		return injectors[i].ImportPath < injectors[j].ImportPath
	})
	targets := make([]Target, 0, len(injectors))
	for _, in := range injectors {
		if opts.Injector != "" && !injectorMatches(in, opts.Injector) {
			continue
		}
		targets = append(targets, Target{
			Name:  in.FuncName,
			Set:   in.Set,
			Roots: []types.Type{in.Out},
		})
	}
	if len(targets) == 0 {
		if opts.Injector != "" {
			return nil, fmt.Errorf("no injector matching %q", opts.Injector)
		}
		return nil, fmt.Errorf("no injectors found; use -set to graph a provider set")
	}
	return targets, nil
}

func injectorMatches(in *wire.Injector, pattern string) bool {
	return in.FuncName == pattern ||
		in.String() == pattern ||
		in.ImportPath+"."+in.FuncName == pattern
}

func providerSetMatches(id wire.ProviderSetID, pattern string) bool {
	return id.VarName == pattern ||
		id.String() == pattern ||
		id.ImportPath+"."+id.VarName == pattern
}

// A builder walks a provider set starting from a set of root types, recording
// the types it reaches and the provider calls that connect them.
type builder struct {
	set   *wire.ProviderSet
	seen  *typeutil.Map // types.Type to *Node
	used  map[[3]string]bool
	nodes []*Node
	edges []*Edge
}

// Build walks t's provider set from its roots and returns the resulting graph.
func Build(t Target) *Graph {
	b := &builder{
		set:  t.Set,
		seen: new(typeutil.Map),
		used: make(map[[3]string]bool),
	}
	b.seen.SetHasher(typeutil.MakeHasher())
	for _, root := range t.Roots {
		b.visit(root)
	}
	return &Graph{Target: t.Name, Nodes: b.nodes, Edges: b.edges}
}

// visit returns the node for t, creating it and its dependencies on first
// visit. Provider sets are acyclic, so the recursion terminates.
func (b *builder) visit(t types.Type) *Node {
	if n := b.seen.At(t); n != nil {
		return n.(*Node)
	}
	pt := b.set.For(t)
	n := b.newNode(t, pt)
	b.seen.Set(t, n)
	switch {
	case pt.IsNil(), pt.IsArg():
		// Injector arguments and unbound inputs have no dependencies.
	case pt.IsProvider():
		p := pt.Provider()
		label := providerEdgeLabel(p)
		for _, arg := range p.Args {
			b.addEdge(b.visit(arg.Type), n, label)
		}
	case pt.IsValue():
		// Values have no dependencies.
	case pt.IsField():
		f := pt.Field()
		b.addEdge(b.visit(f.Parent), n, "field "+f.Name)
	}
	return n
}

func (b *builder) newNode(t types.Type, pt wire.ProvidedType) *Node {
	n := &Node{
		ID:    fmt.Sprintf("n%d", len(b.nodes)),
		Label: shortType(t),
		Pkg:   nodePkg(t, pt),
	}
	switch {
	case pt.IsNil(), pt.IsArg():
		n.Kind = KindInput
	case !types.Identical(t, pt.Type()):
		n.Kind = KindInterface
	case pt.IsValue():
		n.Kind = KindValue
	case pt.IsField():
		n.Kind = KindField
	default:
		n.Kind = KindProvider
	}
	b.nodes = append(b.nodes, n)
	return n
}

func (b *builder) addEdge(from, to *Node, label string) {
	key := [3]string{from.ID, to.ID, label}
	if b.used[key] {
		return
	}
	b.used[key] = true
	b.edges = append(b.edges, &Edge{From: from.ID, To: to.ID, Label: label})
}

// providerEdgeLabel names the provider that connects its arguments to the
// type it produces, qualified by the package the provider is declared in.
func providerEdgeLabel(p *wire.Provider) string {
	name := p.Name
	if len(p.InstanceArgs) > 0 {
		args := make([]string, 0, len(p.InstanceArgs))
		for _, a := range p.InstanceArgs {
			args = append(args, shortType(a))
		}
		name += "[" + strings.Join(args, ", ") + "]"
	}
	if p.Pkg != nil {
		name = p.Pkg.Name() + "." + name
	}
	return name
}

// shortType formats a type with package names instead of full import paths.
func shortType(t types.Type) string {
	return types.TypeString(t, func(p *types.Package) string {
		if p == nil {
			return ""
		}
		return p.Name()
	})
}

// typePkgPath returns the import path of the package that declares t, looking
// through pointers and other type constructors, or "" if t is unnamed.
func typePkgPath(t types.Type) string {
	for {
		switch tt := t.(type) {
		case *types.Pointer:
			t = tt.Elem()
		case *types.Slice:
			t = tt.Elem()
		case *types.Array:
			t = tt.Elem()
		case *types.Chan:
			t = tt.Elem()
		case *types.Named:
			if pkg := tt.Obj().Pkg(); pkg != nil {
				return pkg.Path()
			}
			return ""
		default:
			return ""
		}
	}
}

// nodePkg returns the grouping key for a node.
func nodePkg(t types.Type, pt wire.ProvidedType) string {
	if pkg := typePkgPath(t); pkg != "" {
		return pkg
	}
	switch {
	case pt.IsProvider():
		if pkg := pt.Provider().Pkg; pkg != nil {
			return pkg.Path()
		}
	case pt.IsField():
		if pkg := pt.Field().Pkg; pkg != nil {
			return pkg.Path()
		}
	}
	return "(unnamed types)"
}
