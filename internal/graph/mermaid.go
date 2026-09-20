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
	Register(mermaidRenderer{})
}

// mermaidRenderer writes a Markdown document containing one fenced Mermaid
// diagram per graph.
type mermaidRenderer struct{}

func (mermaidRenderer) Format() string { return "mermaid" }

func (r mermaidRenderer) Render(w io.Writer, graphs []*Graph) error {
	for i, g := range graphs {
		if i > 0 {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "## %s\n\n", g.Target); err != nil {
			return err
		}
		if err := writeMermaid(w, g); err != nil {
			return err
		}
	}
	return nil
}

// writeMermaid writes the graph as a fenced Mermaid diagram.
func writeMermaid(w io.Writer, g *Graph) error {
	ew := &errWriter{w: w}
	fmt.Fprintln(ew, "```mermaid")
	fmt.Fprintln(ew, "flowchart TD")
	for i, grp := range g.Groups() {
		fmt.Fprintf(ew, "    subgraph sg%d[\"%s\"]\n", i, escapeMermaid(grp.Pkg))
		for _, n := range grp.Nodes {
			fmt.Fprintf(ew, "        %s%s\n", n.ID, mermaidNode(n))
		}
		fmt.Fprintln(ew, "    end")
	}
	for _, e := range g.Edges {
		if e.Label == "" {
			fmt.Fprintf(ew, "    %s --> %s\n", e.From, e.To)
			continue
		}
		fmt.Fprintf(ew, "    %s -->|%s| %s\n", e.From, escapeMermaid(e.Label), e.To)
	}
	fmt.Fprintln(ew, "```")
	return ew.err
}

func mermaidNode(n *Node) string {
	label := escapeMermaid(n.Label)
	switch n.Kind {
	case KindInput:
		return fmt.Sprintf(`(["%s"])`, label)
	case KindInterface:
		return fmt.Sprintf(`[["%s"]]`, label)
	case KindValue:
		return fmt.Sprintf(`{{"%s"}}`, label)
	case KindField:
		return fmt.Sprintf(`[/"%s"/]`, label)
	default:
		return fmt.Sprintf(`["%s"]`, label)
	}
}

// escapeMermaid escapes the characters that are special inside a Mermaid
// label using Mermaid's HTML entity codes.
func escapeMermaid(s string) string {
	return strings.NewReplacer(
		"&", "#38;",
		`"`, "#quot;",
		"<", "#60;",
		">", "#62;",
		"|", "#124;",
		"`", "#96;",
		"\n", " ",
	).Replace(s)
}
