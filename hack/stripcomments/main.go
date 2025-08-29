// hack/stripcomments/main.go
package main

import (
	"bufio"
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := strip("."); err != nil {
		panic(err)
	}
}

func strip(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if path != root && (strings.HasPrefix(name, ".") || name == "vendor") {
				return fs.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		return processFile(path)
	})
}

func processFile(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	comment := firstLineComment(src)
	if comment == "" {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, abs)
		if err != nil {
			return err
		}
		comment = "// " + filepath.ToSlash(rel)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, 0)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	buf.WriteString(comment)
	buf.WriteByte('\n')
	if err := printer.Fprint(&buf, fset, file); err != nil {
		return err
	}
	buf.WriteByte('\n')
	return os.WriteFile(path, buf.Bytes(), info.Mode())
}

func firstLineComment(src []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(src))
	if scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "//") {
			return line
		}
	}
	return ""
}
