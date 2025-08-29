// /cmd/commentcheck/main.go
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
		buildLines := 0
		var lines []int
		var texts []string
		for _, cg := range file.Comments {
			for _, c := range cg.List {
				line := fset.Position(c.Slash).Line
				text := strings.TrimSpace(c.Text)
				if strings.HasPrefix(text, "//go:build") || strings.HasPrefix(text, "// +build") {
					if line == buildLines+1 {
						buildLines++
						continue
					}
				}
				lines = append(lines, line)
				texts = append(texts, text)
			}
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		want := "// /" + filepath.ToSlash(rel)
		if len(lines) != 1 || lines[0] != buildLines+1 || len(texts) != 1 || texts[0] != want {
			bad = append(bad, path)
		}
		return nil
	}
	if err := filepath.WalkDir(root, walk); err != nil {
		return err
	}
	if len(bad) > 0 {
		return fmt.Errorf("comment policy violation: %s", strings.Join(bad, ", "))
	}
	return nil
}

func main() {
	if err := check("."); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
