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
	"fmt"
	"io"
	"sort"
	"strings"
)

// A Renderer writes dependency graphs in a particular output format.
//
// To add a format, implement Renderer and call Register from an init
// function; the new format then becomes selectable by name everywhere a
// format is looked up.
type Renderer interface {
	// Format is the name that selects this renderer. It must be lowercase.
	Format() string
	// Render writes graphs to w. Renderers that can only represent a single
	// graph should return an error wrapping ErrMultipleGraphs when given more
	// than one.
	Render(w io.Writer, graphs []*Graph) error
}

// ErrMultipleGraphs is wrapped by the error a Renderer returns when it is
// given more graphs than the format can hold.
var ErrMultipleGraphs = errors.New("format supports only one graph at a time")

var renderers = make(map[string]Renderer)

// Register makes a renderer available under its format name. It panics if the
// name is empty, is not lowercase, or is already registered, so conflicts show
// up at program start. Register is not safe for concurrent use; call it from
// an init function.
func Register(r Renderer) {
	format := r.Format()
	if format == "" || format != strings.ToLower(format) {
		panic(fmt.Sprintf("graph: invalid renderer format %q", format))
	}
	if _, dup := renderers[format]; dup {
		panic(fmt.Sprintf("graph: renderer %q registered twice", format))
	}
	renderers[format] = r
}

// Lookup returns the renderer registered for format, matched
// case-insensitively.
func Lookup(format string) (Renderer, bool) {
	r, ok := renderers[strings.ToLower(format)]
	return r, ok
}

// Formats returns the names of all registered renderers, sorted.
func Formats() []string {
	formats := make([]string, 0, len(renderers))
	for format := range renderers {
		formats = append(formats, format)
	}
	sort.Strings(formats)
	return formats
}

// Render writes graphs to w using the renderer registered for format.
func Render(w io.Writer, format string, graphs []*Graph) error {
	r, ok := Lookup(format)
	if !ok {
		return fmt.Errorf("unknown format %q (want %s)", format, strings.Join(Formats(), " or "))
	}
	return r.Render(w, graphs)
}

// An errWriter lets a renderer write a diagram without checking every call,
// remembering the first error and dropping the writes that follow it.
type errWriter struct {
	w   io.Writer
	err error
}

func (ew *errWriter) Write(p []byte) (int, error) {
	if ew.err != nil {
		return len(p), nil
	}
	n, err := ew.w.Write(p)
	ew.err = err
	return n, err
}
