// cmd/commentcheck/main.go
package main

import (
	"bufio"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var dirs = []string{"cli", "cmd/commentcheck", "config", "internal", "patternmatching"}

func main() {
	var files []string
	entries, err := os.ReadDir(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") {
			files = append(files, e.Name())
		}
	}
	for _, d := range dirs {
		err := filepath.WalkDir(d, func(path string, de fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if de.IsDir() {
				return nil
			}
			if filepath.Ext(path) == ".go" {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	var failed bool
	for _, f := range files {
		if err := checkFile(f); err != nil {
			fmt.Fprintln(os.Stderr, err)
			failed = true
		}
	}
	if failed {
		os.Exit(1)
	}
}

func checkFile(path string) error {
	rel := filepath.ToSlash(path)
	expected := "// " + rel
	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fh.Close()
	reader := bufio.NewReader(fh)
	first, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("%s: unable to read first line", path)
	}
	first = strings.TrimRight(first, "\n")
	if first != expected {
		return fmt.Errorf("%s: first line must be %q", path, expected)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("%s: %v", path, err)
	}
	if len(file.Comments) > 1 {
		return fmt.Errorf("%s: additional comments found", path)
	}
	if len(file.Comments) == 0 {
		return fmt.Errorf("%s: missing file comment", path)
	}
	firstGroup := file.Comments[0]
	pos := fset.Position(firstGroup.Pos())
	if pos.Line != 1 || len(firstGroup.List) != 1 || firstGroup.List[0].Text != expected {
		return fmt.Errorf("%s: first comment must be %q", path, expected)
	}
	return nil
}
