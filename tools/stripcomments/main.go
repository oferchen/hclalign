// filename: tools/stripcomments/main.go
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func strip(node ast.Node) {
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.File:
			x.Doc = nil
		case *ast.GenDecl:
			x.Doc = nil
		case *ast.FuncDecl:
			x.Doc = nil
		case *ast.Field:
			x.Doc, x.Comment = nil, nil
		case *ast.ImportSpec:
			x.Doc, x.Comment = nil, nil
		case *ast.TypeSpec:
			x.Doc, x.Comment = nil, nil
		case *ast.ValueSpec:
			x.Doc, x.Comment = nil, nil
		}
		return true
	})
}

func process(root, path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return err
	}
	src, err := os.ReadFile(abs)
	if err != nil {
		return err
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, abs, src, parser.ParseComments)
	if err != nil {
		return err
	}
	var tags []string
	for _, cg := range f.Comments {
		if cg.Pos() > f.Package {
			continue
		}
		for _, c := range cg.List {
			t := strings.TrimSpace(c.Text)
			if strings.HasPrefix(t, "//go:build") || strings.HasPrefix(t, "// +build") {
				tags = append(tags, t)
			}
		}
	}
	f.Comments = nil
	strip(f)
	var buf bytes.Buffer
	if err := (&printer.Config{Mode: printer.TabIndent | printer.UseSpaces}).Fprint(&buf, fset, f); err != nil {
		return err
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	for _, t := range tags {
		out.WriteString(t)
		out.WriteByte('\n')
	}
	if len(tags) > 0 {
		out.WriteByte('\n')
	}
	out.WriteString("// filename: " + filepath.ToSlash(rel) + "\n")
	out.Write(formatted)
	b := out.Bytes()
	if len(b) == 0 || b[len(b)-1] != '\n' {
		out.WriteByte('\n')
		b = out.Bytes()
	}
	return os.WriteFile(path, b, info.Mode())
}

func main() {
	var repoRoot string
	flag.StringVar(&repoRoot, "repo-root", "", "repository root")
	flag.Parse()
	if repoRoot == "" {
		fmt.Fprintln(os.Stderr, "--repo-root is required")
		os.Exit(1)
	}
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if d.IsDir() && rel == filepath.Join("tools", "stripcomments") {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}
		return process(root, path)
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
