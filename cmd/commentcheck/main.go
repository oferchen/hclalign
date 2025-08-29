// cmd/commentcheck/main.go
package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func check(root string) error {
	var bad []string
	walk := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "vendor" || strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		fset := token.NewFileSet()
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
		if err != nil {
			return err
		}
		var nonBuild []token.Position
		for _, cg := range file.Comments {
			for _, c := range cg.List {
				if strings.HasPrefix(c.Text, "//go:build") {
					continue
				}
				nonBuild = append(nonBuild, fset.Position(c.Slash))
			}
		}
		if len(nonBuild) != 1 || nonBuild[0].Line != 1 {
			bad = append(bad, path)
		}
		return nil
	}
	if err := filepath.WalkDir(root, walk); err != nil {
		return err
	}
	if len(bad) > 0 {
		return fmt.Errorf("comments beyond first line: %s", strings.Join(bad, ", "))
	}
	return nil
}

func main() {
	if err := check("."); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
