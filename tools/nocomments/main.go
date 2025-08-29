// tools/nocomments/main.go
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
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
	out.WriteString("// " + filepath.ToSlash(rel) + "\n")
	out.Write(formatted)
	b := out.Bytes()
	if len(b) == 0 || b[len(b)-1] != '\n' {
		out.WriteByte('\n')
		b = out.Bytes()
	}
	return os.WriteFile(path, b, info.Mode())
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(os.Args) < 2 {
		return
	}
	for _, p := range os.Args[1:] {
		if err := process(root, p); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
