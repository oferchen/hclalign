// cmd/commentcheck/main.go
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	execCommand = exec.Command
	lookPath    = exec.LookPath
)

func main() {
	dirs, err := packageDirs()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var files []string
	for _, d := range dirs {
		entries, err := os.ReadDir(d)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if strings.HasSuffix(e.Name(), ".go") {
				files = append(files, filepath.Join(d, e.Name()))
			}
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

func packageDirs() ([]string, error) {
	if _, err := lookPath("go"); err != nil {
		return nil, fmt.Errorf("commentcheck requires a Go toolchain: %w", err)
	}
	out, err := execCommand("go", "list", "-f", "{{.Dir}}", "./...").Output()
	if err != nil {
		return nil, err
	}
	var dirs []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		dir := scanner.Text()
		rel, err := filepath.Rel(".", dir)
		if err != nil {
			return nil, err
		}
		dirs = append(dirs, rel)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return dirs, nil
}
