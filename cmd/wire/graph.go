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

package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/subcommands"
	"github.com/wireinject/wire/internal/graph"
	"github.com/wireinject/wire/internal/wire"
)

type graphCmd struct {
	format   string
	injector string
	set      string
	tags     string
	output   string
}

func (*graphCmd) Name() string { return "graph" }
func (*graphCmd) Synopsis() string {
	return "output the dependency graph of injectors or provider sets"
}
func (*graphCmd) Usage() string {
	return fmt.Sprintf(`graph [flags] [packages]

  Given one or more packages, graph outputs the dependency graph of each
  injector function defined in them, or of a top-level provider set with -set.

  The default output is a Markdown document containing a Mermaid diagram.
  Use -format=dot for Graphviz DOT output. The available formats are %s.

  If no packages are listed, it defaults to ".".
`, strings.Join(graph.Formats(), ", "))
}
func (cmd *graphCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.format, "format", "mermaid", "output format: "+strings.Join(graph.Formats(), " or "))
	f.StringVar(&cmd.injector, "injector", "", "only graph the injector with this name (unqualified or quoted qualified name)")
	f.StringVar(&cmd.set, "set", "", "graph this top-level provider set instead of injectors (variable name or quoted qualified name)")
	f.StringVar(&cmd.tags, "tags", "", "append build tags to the default wirebuild")
	f.StringVar(&cmd.output, "output", "", "write the graph to this file instead of stdout")
}

func (cmd *graphCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	renderer, ok := graph.Lookup(cmd.format)
	if !ok {
		log.Printf("unknown format %q (want %s)\n", cmd.format, strings.Join(graph.Formats(), " or "))
		return subcommands.ExitUsageError
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Println("failed to get working directory: ", err)
		return subcommands.ExitFailure
	}
	info, errs := wire.Load(ctx, wd, os.Environ(), cmd.tags, packages(f))
	if len(errs) > 0 {
		logErrors(errs)
		log.Println("error loading packages")
		return subcommands.ExitFailure
	}
	if info == nil {
		log.Println("error loading packages")
		return subcommands.ExitFailure
	}
	targets, err := graph.Targets(info, graph.Options{Injector: cmd.injector, Set: cmd.set})
	if err != nil {
		log.Println(err)
		return subcommands.ExitFailure
	}
	graphs := make([]*graph.Graph, 0, len(targets))
	for _, t := range targets {
		graphs = append(graphs, graph.Build(t))
	}
	var buf bytes.Buffer
	if err := renderer.Render(&buf, graphs); err != nil {
		if errors.Is(err, graph.ErrMultipleGraphs) {
			log.Printf("%v; use -injector or -set to select one\n", err)
		} else {
			log.Println(err)
		}
		return subcommands.ExitFailure
	}
	if cmd.output != "" {
		if err := os.WriteFile(cmd.output, buf.Bytes(), 0644); err != nil {
			log.Printf("failed to write %s: %v\n", cmd.output, err)
			return subcommands.ExitFailure
		}
		log.Printf("wrote %s\n", cmd.output)
		return subcommands.ExitSuccess
	}
	os.Stdout.Write(buf.Bytes())
	return subcommands.ExitSuccess
}
