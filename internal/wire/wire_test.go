// Copyright 2018 The Wire Authors
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

package wire

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"go/build"
	"go/types"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/google/go-cmp/cmp"
)

var record = flag.Bool("record", false, "whether to run tests against cloud resources and record the interactions")

func TestWire(t *testing.T) {
	const testRoot = "testdata"
	testdataEnts, err := ioutil.ReadDir(testRoot) // ReadDir sorts by name.
	if err != nil {
		t.Fatal(err)
	}
	// The marker function package source is needed to have the test cases
	// type check. loadTestCase places this file at the well-known import path.
	wireGo, err := ioutil.ReadFile(filepath.Join("..", "..", "wire.go"))
	if err != nil {
		t.Fatal(err)
	}
	tests := make([]*testCase, 0, len(testdataEnts))
	for _, ent := range testdataEnts {
		name := ent.Name()
		if !ent.IsDir() || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
			continue
		}
		test, err := loadTestCase(filepath.Join(testRoot, name), wireGo)
		if err != nil {
			t.Error(err)
			continue
		}
		tests = append(tests, test)
	}

	var goToolPath string
	if *record {
		goToolPath = filepath.Join(build.Default.GOROOT, "bin", "go")
		if _, err := os.Stat(goToolPath); err != nil {
			t.Fatal("go toolchain not available:", err)
		}
	}
	ctx := context.Background()
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Materialize a temporary GOPATH directory.
			gopath, err := ioutil.TempDir("", "wire_test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(gopath)
			gopath, err = filepath.EvalSymlinks(gopath)
			if err != nil {
				t.Fatal(err)
			}
			if err := test.materialize(gopath); err != nil {
				t.Fatal(err)
			}
			wd := filepath.Join(gopath, "src", "example.com")
			gens, errs := Generate(ctx, wd, append(os.Environ(), "GOPATH="+gopath), []string{test.pkg}, &GenerateOptions{Header: test.header})
			var gen GenerateResult
			if len(gens) > 1 {
				t.Fatalf("got %d generated files, want 0 or 1", len(gens))
			}
			if len(gens) == 1 {
				gen = gens[0]
				if len(gen.Errs) > 0 {
					errs = append(errs, gen.Errs...)
				}
				if len(gen.Content) > 0 {
					defer t.Logf("wire_gen.go:\n%s", gen.Content)
				}
			}
			if len(errs) > 0 {
				gotErrStrings := make([]string, len(errs))
				for i, e := range errs {
					t.Log(e.Error())
					gotErrStrings[i] = scrubError(gopath, e.Error())
				}
				if !test.wantWireError {
					t.Fatal("Did not expect errors. To -record an error, create want/wire_errs.txt.")
				}
				if *record {
					wireErrsFile := filepath.Join(testRoot, test.name, "want", "wire_errs.txt")
					if err := ioutil.WriteFile(wireErrsFile, []byte(strings.Join(gotErrStrings, "\n\n")), 0666); err != nil {
						t.Fatalf("failed to write wire_errs.txt file: %v", err)
					}
				} else {
					if diff := cmp.Diff(gotErrStrings, test.wantWireErrorStrings); diff != "" {
						t.Errorf("Errors didn't match expected errors from wire_errors.txt:\n%s", diff)
					}
				}
				return
			}
			if test.wantWireError {
				t.Fatal("wire succeeded; want error")
			}
			outPathSane := true
			if prefix := gopath + string(os.PathSeparator) + "src" + string(os.PathSeparator); !strings.HasPrefix(gen.OutputPath, prefix) {
				outPathSane = false
				t.Errorf("suggested output path = %q; want to start with %q", gen.OutputPath, prefix)
			}

			if *record {
				// Record ==> Build the generated Wire code,
				// check that the program's output matches the
				// expected output, save wire output on
				// success.
				if !outPathSane {
					return
				}
				if err := gen.Commit(); err != nil {
					t.Fatalf("failed to write wire_gen.go to test GOPATH: %v", err)
				}
				if err := goBuildCheck(goToolPath, gopath, test); err != nil {
					t.Fatalf("go build check failed: %v", err)
				}
				testdataWireGenPath := filepath.Join(testRoot, test.name, "want", "wire_gen.go")
				if err := ioutil.WriteFile(testdataWireGenPath, gen.Content, 0666); err != nil {
					t.Fatalf("failed to record wire_gen.go to testdata: %v", err)
				}
			} else {
				// Replay ==> Load golden file and compare to
				// generated result. This check is meant to
				// detect non-deterministic behavior in the
				// Generate function.
				if !bytes.Equal(gen.Content, test.wantWireOutput) {
					gotS, wantS := string(gen.Content), string(test.wantWireOutput)
					diff := cmp.Diff(strings.Split(gotS, "\n"), strings.Split(wantS, "\n"))
					t.Fatalf("wire output differs from golden file. If this change is expected, run with -record to update the wire_gen.go file.\n*** got:\n%s\n\n*** want:\n%s\n\n*** diff:\n%s", gotS, wantS, diff)
				}
			}
		})
	}
}

func goBuildCheck(goToolPath, gopath string, test *testCase) error {
	// Run `go build`.
	// `go build -o` uses the output path verbatim and does not append the
	// executable suffix, so add it ourselves on Windows.
	testExePath := filepath.Join(gopath, "bin", "testprog")
	if runtime.GOOS == "windows" {
		testExePath += ".exe"
	}
	buildCmd := []string{"build", "-o", testExePath}
	buildCmd = append(buildCmd, test.pkg)
	cmd := exec.Command(goToolPath, buildCmd...)
	cmd.Dir = filepath.Join(gopath, "src", "example.com")
	cmd.Env = append(os.Environ(), "GOPATH="+gopath)
	if buildOut, err := cmd.CombinedOutput(); err != nil {
		if len(buildOut) > 0 {
			return fmt.Errorf("build: %v; output:\n%s", err, buildOut)
		}
		return fmt.Errorf("build: %v", err)
	}

	// Run the resulting program and compare its output to the expected
	// output. Surface stderr on failure so panics are visible.
	out, err := exec.Command(testExePath).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && len(ee.Stderr) > 0 {
			return fmt.Errorf("run compiled program: %v\n%s", err, ee.Stderr)
		}
		return fmt.Errorf("run compiled program: %v", err)
	}
	if !bytes.Equal(out, test.wantProgramOutput) {
		gotS, wantS := string(out), string(test.wantProgramOutput)
		diff := cmp.Diff(strings.Split(gotS, "\n"), strings.Split(wantS, "\n"))
		return fmt.Errorf("compiled program output doesn't match:\n*** got:\n%s\n\n*** want:\n%s\n\n*** diff:\n%s", gotS, wantS, diff)
	}
	return nil
}

func TestUnexport(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"", ""},
		{"a", "a"},
		{"ab", "ab"},
		{"A", "a"},
		{"AB", "ab"},
		{"A_", "a_"},
		{"ABc", "aBc"},
		{"ABC", "abc"},
		{"AB_", "ab_"},
		{"foo", "foo"},
		{"Foo", "foo"},
		{"HTTPClient", "httpClient"},
		{"IFace", "iFace"},
		{"SNAKE_CASE", "snake_CASE"},
		{"HTTP", "http"},
	}
	for _, test := range tests {
		if got := unexport(test.name); got != test.want {
			t.Errorf("unexport(%q) = %q; want %q", test.name, got, test.want)
		}
	}
}

func TestExport(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"", ""},
		{"a", "A"},
		{"ab", "Ab"},
		{"A", "A"},
		{"AB", "AB"},
		{"A_", "A_"},
		{"ABc", "ABc"},
		{"ABC", "ABC"},
		{"AB_", "AB_"},
		{"foo", "Foo"},
		{"Foo", "Foo"},
		{"HTTPClient", "HTTPClient"},
		{"httpClient", "HttpClient"},
		{"IFace", "IFace"},
		{"iFace", "IFace"},
		{"SNAKE_CASE", "SNAKE_CASE"},
		{"HTTP", "HTTP"},
	}
	for _, test := range tests {
		if got := export(test.name); got != test.want {
			t.Errorf("export(%q) = %q; want %q", test.name, got, test.want)
		}
	}
}

func TestTypeVariableName(t *testing.T) {
	var (
		boolT           = types.Typ[types.Bool]
		stringT         = types.Typ[types.String]
		fooVarT         = types.NewNamed(types.NewTypeName(0, nil, "foo", stringT), stringT, nil)
		nonameVarT      = types.NewNamed(types.NewTypeName(0, nil, "", stringT), stringT, nil)
		barVarInFooPkgT = types.NewNamed(types.NewTypeName(0, types.NewPackage("my.example/foo", "foo"), "bar", stringT), stringT, nil)
	)
	tests := []struct {
		description     string
		typ             types.Type
		defaultName     string
		transformAppend string
		collides        map[string]bool
		want            string
	}{
		{"basic type", boolT, "", "", map[string]bool{}, "bool"},
		{"basic type with transform", boolT, "", "suffix", map[string]bool{}, "boolsuffix"},
		{"basic type with collision", boolT, "", "", map[string]bool{"bool": true}, "bool2"},
		{"basic type with transform and collision", boolT, "", "suffix", map[string]bool{"boolsuffix": true}, "boolsuffix2"},
		{"a different basic type", stringT, "", "", map[string]bool{}, "string"},
		{"named type", fooVarT, "", "", map[string]bool{}, "foo"},
		{"named type with transform", fooVarT, "", "suffix", map[string]bool{}, "foosuffix"},
		{"named type with collision", fooVarT, "", "", map[string]bool{"foo": true}, "foo2"},
		{"named type with transform and collision", fooVarT, "", "suffix", map[string]bool{"foosuffix": true}, "foosuffix2"},
		{"noname type", nonameVarT, "bar", "", map[string]bool{}, "bar"},
		{"noname type with transform", nonameVarT, "bar", "s", map[string]bool{}, "bars"},
		{"noname type with transform and collision", nonameVarT, "bar", "s", map[string]bool{"bars": true}, "bars2"},
		{"var in pkg type", barVarInFooPkgT, "", "", map[string]bool{}, "bar"},
		{"var in pkg type with collision", barVarInFooPkgT, "", "", map[string]bool{"bar": true}, "fooBar"},
		{"var in pkg type with double collision", barVarInFooPkgT, "", "", map[string]bool{"bar": true, "fooBar": true}, "bar2"},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s: typeVariableName(%v, %q, %q, %v)", test.description, test.typ, test.defaultName, test.transformAppend, test.collides), func(t *testing.T) {
			got := typeVariableName(test.typ, test.defaultName, func(name string) string { return name + test.transformAppend }, func(name string) bool { return test.collides[name] })
			if !isIdent(got) {
				t.Errorf("%q is not an identifier", got)
			}
			if got != test.want {
				t.Errorf("got %q want %q", got, test.want)
			}
			if test.collides[got] {
				t.Errorf("%q collides", got)
			}
		})
	}
}

func TestDisambiguate(t *testing.T) {
	tests := []struct {
		name     string
		want     string
		collides map[string]bool
	}{
		{"foo", "foo", nil},
		{"foo", "foo2", map[string]bool{"foo": true}},
		{"foo", "foo3", map[string]bool{"foo": true, "foo1": true, "foo2": true}},
		{"foo1", "foo1_2", map[string]bool{"foo": true, "foo1": true, "foo2": true}},
		{"foo\u0661", "foo\u0661", map[string]bool{"foo": true, "foo1": true, "foo2": true}},
		{"foo\u0661", "foo\u06612", map[string]bool{"foo": true, "foo1": true, "foo2": true, "foo\u0661": true}},
		{"select", "select2", nil},
		{"var", "var2", nil},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("disambiguate(%q, %v)", test.name, test.collides), func(t *testing.T) {
			got := disambiguate(test.name, func(name string) bool { return test.collides[name] })
			if !isIdent(got) {
				t.Errorf("%q is not an identifier", got)
			}
			if got != test.want {
				t.Errorf("got %q want %q", got, test.want)
			}
			if test.collides[got] {
				t.Errorf("%q collides", got)
			}
		})
	}
}

func isIdent(s string) bool {
	if len(s) == 0 {
		return false
	}
	r, i := utf8.DecodeRuneInString(s)
	if !unicode.IsLetter(r) && r != '_' {
		return false
	}
	for i < len(s) {
		r, sz := utf8.DecodeRuneInString(s[i:])
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
		i += sz
	}
	return true
}

// scrubError rewrites the given string to remove occurrences of GOPATH/src,
// rewrites OS-specific path separators to slashes, and any line/column
// information to a fixed ":x:y". For example, if the gopath parameter is
// "C:\GOPATH" and running on Windows, the string
// "C:\GOPATH\src\foo\bar.go:15:4" would be rewritten to "foo/bar.go:x:y".
func scrubError(gopath string, s string) string {
	// Go versions differ on the optional "name" prefix in this diagnostic.
	if strings.Contains(s, " not exported by package ") {
		s = strings.ReplaceAll(s, ": name ", ": ")
	}
	sb := new(strings.Builder)
	query := gopath + string(os.PathSeparator) + "src" + string(os.PathSeparator)
	for {
		// Find next occurrence of source root. This indicates the next path to
		// scrub.
		start := strings.Index(s, query)
		if start == -1 {
			sb.WriteString(s)
			break
		}

		// Find end of file name (extension ".go").
		fileStart := start + len(query)
		fileEnd := strings.Index(s[fileStart:], ".go")
		if fileEnd == -1 {
			// If no ".go" occurs to end of string, further searches will fail too.
			// Break the loop.
			sb.WriteString(s)
			break
		}
		fileEnd += fileStart + 3 // Advance to end of extension.

		// Write out file name and advance scrub position.
		file := s[fileStart:fileEnd]
		if os.PathSeparator != '/' {
			file = strings.Replace(file, string(os.PathSeparator), "/", -1)
		}
		sb.WriteString(s[:start])
		sb.WriteString(file)
		s = s[fileEnd:]

		// Peek past to see if there is line/column info.
		linecol, linecolLen := scrubLineColumn(s)
		sb.WriteString(linecol)
		s = s[linecolLen:]
	}
	return sb.String()
}

func scrubLineColumn(s string) (replacement string, n int) {
	if !strings.HasPrefix(s, ":") {
		return "", 0
	}
	// Skip first colon and run of digits.
	for n++; len(s) > n && '0' <= s[n] && s[n] <= '9'; {
		n++
	}
	if n == 1 {
		// No digits followed colon.
		return "", 0
	}

	// Start on column part.
	if !strings.HasPrefix(s[n:], ":") {
		return ":x", n
	}
	lineEnd := n
	// Skip second colon and run of digits.
	for n++; len(s) > n && '0' <= s[n] && s[n] <= '9'; {
		n++
	}
	if n == lineEnd+1 {
		// No digits followed second colon.
		return ":x", lineEnd
	}
	return ":x:y", n
}

type testCase struct {
	name                 string
	pkg                  string
	header               []byte
	goFiles              map[string][]byte
	wantProgramOutput    []byte
	wantWireOutput       []byte
	wantWireError        bool
	wantWireErrorStrings []string
}

// loadTestCase reads a test case from a directory.
//
// The directory structure is:
//
//	root/
//
//		pkg
//			file containing the package name containing the inject function
//			(must also be package main)
//
//		...
//			any Go files found recursively placed under GOPATH/src/...
//
//		want/
//
//			wire_errs.txt
//					Expected errors from the Wire Generate function,
//					missing if no errors expected.
//					Distinct errors are separated by a blank line,
//					and line numbers and line positions are scrubbed
//					(e.g. "$GOPATH/src/foo.go:52:8" --> "foo.go:x:y").
//
//			wire_gen.go
//					verified output of wire from a test run with
//					-record, missing if wire_errs.txt is present
//
//			program_out.txt
//					expected output from the final compiled program,
//					missing if wire_errs.txt is present
func loadTestCase(root string, wireGoSrc []byte) (*testCase, error) {
	name := filepath.Base(root)
	pkg, err := ioutil.ReadFile(filepath.Join(root, "pkg"))
	if err != nil {
		return nil, fmt.Errorf("load test case %s: %v", name, err)
	}
	header, _ := ioutil.ReadFile(filepath.Join(root, "header"))
	var wantProgramOutput []byte
	var wantWireOutput []byte
	wireErrb, err := ioutil.ReadFile(filepath.Join(root, "want", "wire_errs.txt"))
	wantWireError := err == nil
	var wantWireErrorStrings []string
	if wantWireError {
		for _, errs := range strings.Split(string(wireErrb), "\n\n") {
			// Allow for trailing newlines, which can be hard to remove in some editors.
			wantWireErrorStrings = append(wantWireErrorStrings, strings.TrimRight(errs, "\n\r"))
		}
	} else {
		if !*record {
			wantWireOutput, err = ioutil.ReadFile(filepath.Join(root, "want", "wire_gen.go"))
			if err != nil {
				return nil, fmt.Errorf("load test case %s: %v, if this is a new testcase, run with -record to generate the wire_gen.go file", name, err)
			}
		}
		wantProgramOutput, err = ioutil.ReadFile(filepath.Join(root, "want", "program_out.txt"))
		if err != nil {
			return nil, fmt.Errorf("load test case %s: %v", name, err)
		}
	}
	goFiles := map[string][]byte{
		"github.com/google/wire/wire.go": wireGoSrc,
	}
	err = filepath.Walk(root, func(src string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, src)
		if err != nil {
			return err // unlikely
		}
		if info.Mode().IsDir() && rel == "want" {
			// The "want" directory should not be included in goFiles.
			return filepath.SkipDir
		}
		if !info.Mode().IsRegular() || filepath.Ext(src) != ".go" {
			return nil
		}
		data, err := ioutil.ReadFile(src)
		if err != nil {
			return err
		}
		goFiles["example.com/"+filepath.ToSlash(rel)] = data
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("load test case %s: %v", name, err)
	}
	wantWireOutput = []byte(strings.Replace(string(wantWireOutput), "//go:generate go run -mod=mod github.com/google/wire/cmd/wire", "//go:generate go run -mod=mod github.com/wireinject/wire/cmd/wire", 1))
	return &testCase{
		name:                 name,
		pkg:                  string(bytes.TrimSpace(pkg)),
		header:               header,
		goFiles:              goFiles,
		wantWireOutput:       wantWireOutput,
		wantProgramOutput:    wantProgramOutput,
		wantWireError:        wantWireError,
		wantWireErrorStrings: wantWireErrorStrings,
	}, nil
}

// materialize creates a new GOPATH at the given directory, which may or
// may not exist.
func (test *testCase) materialize(gopath string) error {
	for name, content := range test.goFiles {
		dst := filepath.Join(gopath, "src", filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
			return fmt.Errorf("materialize GOPATH: %v", err)
		}
		if err := ioutil.WriteFile(dst, content, 0666); err != nil {
			return fmt.Errorf("materialize GOPATH: %v", err)
		}
	}

	// Add go.mod files to example.com and github.com/google/wire.
	const importPath = "example.com"
	const depPath = "github.com/google/wire"
	depLoc := filepath.Join(gopath, "src", filepath.FromSlash(depPath))
	example := fmt.Sprintf("module %s\n\ngo 1.19\n\nrequire %s v0.1.0\nreplace %s => %s\n", importPath, depPath, depPath, depLoc)
	gomod := filepath.Join(gopath, "src", filepath.FromSlash(importPath), "go.mod")
	if err := ioutil.WriteFile(gomod, []byte(example), 0666); err != nil {
		return fmt.Errorf("generate go.mod for %s: %v", gomod, err)
	}
	if err := ioutil.WriteFile(filepath.Join(depLoc, "go.mod"), []byte("module "+depPath+"\n"), 0666); err != nil {
		return fmt.Errorf("generate go.mod for %s: %v", depPath, err)
	}
	return nil
}

// TestWireNilCleanup compiles and runs generated injectors that mix nil and
// non-nil provider cleanups. Golden snapshots cannot catch a nil func panic.
func TestWireNilCleanup(t *testing.T) {
	marker, err := os.ReadFile(filepath.Join("..", "..", "wire.go"))
	if err != nil {
		t.Fatal(err)
	}
	const source = `//go:build wireinject

package main

import (
	"errors"
	"fmt"
	"github.com/google/wire"
)

const fail = %t
const allNil = %t

var cleaned []string
var providerErr = errors.New("provider failed")

type A struct{}
type B struct{}
type C struct{}
type Result struct{ Ready bool }

func cleanupFor(name string) func() {
	if allNil {
		return nil
	}
	return func() { cleaned = append(cleaned, name) }
}

func provideA() (A, func()) { return A{}, cleanupFor("A") }
func provideB(A) (B, func()) { return B{}, nil }
func provideC(B) (C, func()) { return C{}, cleanupFor("C") }
func provideResult(C) (Result, func(), error) {
	if fail {
		return Result{Ready: true}, func() { panic("failed provider cleanup called") }, providerErr
	}
	return Result{Ready: true}, nil, nil
}

func inject() (Result, func(), error) {
	wire.Build(provideA, provideB, provideC, provideResult)
	return Result{}, nil, nil
}

func main() {
	result, cleanup, err := inject()
	if fail {
		// Identity: the injector must return the provider error, not a wrapper.
		if err != providerErr {
			panic("error is not the provider error")
		}
		if result != (Result{}) {
			panic("result is not the zero value")
		}
		if cleanup != nil {
			panic("aggregate cleanup is non-nil")
		}
	} else {
		if err != nil {
			panic("success path returned an error")
		}
		if !result.Ready {
			panic("result not ready")
		}
		if cleanup == nil {
			panic("aggregate cleanup is nil")
		}
		if len(cleaned) != 0 {
			panic("providers cleaned before the aggregate cleanup ran")
		}
		cleanup()
	}
	fmt.Println(cleaned)
}
`
	for _, tc := range []struct {
		name   string
		fail   bool
		allNil bool
		want   string
	}{
		{name: "success", want: "[C A]\n"},
		{name: "error", fail: true, want: "[C A]\n"},
		{name: "all_nil", allNil: true, want: "[]\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			test := &testCase{
				pkg: "example.com/foo",
				goFiles: map[string][]byte{
					"github.com/google/wire/wire.go": marker,
					"example.com/foo/foo.go":         []byte(fmt.Sprintf(source, tc.fail, tc.allNil)),
				},
				wantProgramOutput: []byte(tc.want),
			}
			gopath := t.TempDir()
			if err := test.materialize(gopath); err != nil {
				t.Fatal(err)
			}
			gens, errs := Generate(context.Background(), filepath.Join(gopath, "src", "example.com"),
				append(os.Environ(), "GOPATH="+gopath), []string{test.pkg}, nil)
			if len(errs) != 0 || len(gens) != 1 {
				t.Fatalf("Generate: %d results, errors: %v", len(gens), errs)
			}
			if len(gens[0].Errs) != 0 {
				t.Fatal(gens[0].Errs)
			}
			if err := gens[0].Commit(); err != nil {
				t.Fatal(err)
			}
			if err := goBuildCheck(filepath.Join(build.Default.GOROOT, "bin", "go"), gopath, test); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Pointer aliases have different go/types representations across Go releases.
// Check the compiled program rather than fixing one representation in a golden.
func TestWirePointerAliases(t *testing.T) {
	marker, err := os.ReadFile(filepath.Join("..", "..", "wire.go"))
	if err != nil {
		t.Fatal(err)
	}
	test := &testCase{
		pkg: "example.com/foo",
		goFiles: map[string][]byte{
			"github.com/google/wire/wire.go": marker,
			"example.com/foo/foo.go": []byte(`//go:build wireinject

package main

import (
	"fmt"
	"github.com/google/wire"
)

type S struct { N int }
type Pointer = *S
var marker Pointer

func injectPointer() Pointer {
	wire.Build(wire.Value(7), wire.Struct(new(S), "*"))
	return nil
}
func injectMarker() S {
	wire.Build(wire.Value(9), wire.Struct(marker, "*"))
	return S{}
}
func main() {
	if injectPointer().N != 7 || injectMarker().N != 9 {
		panic("incorrect pointer alias injection")
	}
	fmt.Println("ok")
}
`),
		},
		wantProgramOutput: []byte("ok\n"),
	}
	gopath := t.TempDir()
	if err := test.materialize(gopath); err != nil {
		t.Fatal(err)
	}
	gens, errs := Generate(context.Background(), filepath.Join(gopath, "src", "example.com"),
		append(os.Environ(), "GOPATH="+gopath), []string{test.pkg}, nil)
	if len(errs) != 0 || len(gens) != 1 {
		t.Fatalf("Generate: %d results, errors: %v", len(gens), errs)
	}
	if len(gens[0].Errs) != 0 {
		t.Fatal(gens[0].Errs)
	}
	if err := gens[0].Commit(); err != nil {
		t.Fatal(err)
	}
	if err := goBuildCheck(filepath.Join(build.Default.GOROOT, "bin", "go"), gopath, test); err != nil {
		t.Fatal(err)
	}
}
