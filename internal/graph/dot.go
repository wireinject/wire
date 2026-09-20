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
)

func init() {
	Register(dotRenderer{})
}

// dotRenderer writes a graph in Graphviz DOT format. A DOT file holds a
// single digraph, so it renders one graph at a time.
type dotRenderer struct{}

func (dotRenderer) Format() string { return "dot" }

func (r dotRenderer) Render(w io.Writer, graphs []*Graph) error {
	if len(graphs) > 1 {
		return fmt.Errorf("dot output got %d graphs: %w", len(graphs), ErrMultipleGraphs)
	}
	for _, g := range graphs {
		if err := writeDOT(w, g); err != nil {
			return err
		}
	}
	return nil
}

// writeDOT writes the graph in Graphviz DOT format.
func writeDOT(w io.Writer, g *Graph) error {
	ew := &errWriter{w: w}
	fmt.Fprintf(ew, "digraph %s {\n", dotQuote(g.Target))
	fmt.Fprintln(ew, "    rankdir=TB;")
	fmt.Fprintln(ew, "    node [shape=box];")
	for i, grp := range g.Groups() {
		fmt.Fprintf(ew, "    subgraph cluster%d {\n", i)
		fmt.Fprintf(ew, "        label=%s;\n", dotQuote(grp.Pkg))
		for _, n := range grp.Nodes {
			fmt.Fprintf(ew, "        %s;\n", dotQuote(n.ID))
		}
		fmt.Fprintln(ew, "    }")
	}
	for _, n := range g.Nodes {
		fmt.Fprintf(ew, "    %s [label=%s%s];\n", dotQuote(n.ID), dotQuote(n.Label), dotShape(n.Kind))
	}
	for _, e := range g.Edges {
		if e.Label == "" {
			fmt.Fprintf(ew, "    %s -> %s;\n", dotQuote(e.From), dotQuote(e.To))
			continue
		}
		fmt.Fprintf(ew, "    %s -> %s [label=%s];\n", dotQuote(e.From), dotQuote(e.To), dotQuote(e.Label))
	}
	fmt.Fprintln(ew, "}")
	return ew.err
}

func dotShape(kind Kind) string {
	switch kind {
	case KindInput:
		return ", shape=ellipse"
	case KindInterface:
		return ", peripheries=2"
	case KindValue:
		return ", shape=hexagon"
	case KindField:
		return ", shape=parallelogram"
	default:
		return ""
	}
}

func dotQuote(s string) string {
	return `"` + strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
		"\n", `\n`,
	).Replace(s) + `"`
}
