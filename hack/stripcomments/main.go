// hack/stripcomments/main.go — SPDX-License-Identifier: Apache-2.0
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
	comment := "// " + filepath.ToSlash(rel) + " — SPDX-License-Identifier: Apache-2.0"

	tags := extractBuildTags(src)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, 0)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	for _, t := range tags {
		buf.WriteString(t)
		buf.WriteByte('\n')
	}
	if len(tags) > 0 {
		buf.WriteByte('\n')
	}
	buf.WriteString(comment)
	buf.WriteByte('\n')
	if err := printer.Fprint(&buf, fset, file); err != nil {
		return err
	}
	buf.WriteByte('\n')
	return os.WriteFile(path, buf.Bytes(), info.Mode())
}

func extractBuildTags(src []byte) []string {
	scanner := bufio.NewScanner(bytes.NewReader(src))
	var tags []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "//go:build") {
			tags = append(tags, line)
			continue
		}
		if strings.TrimSpace(line) == "" && len(tags) > 0 {
			break
		}
		break
	}
	return tags
}
