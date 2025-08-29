// cmd/commentcheck/main.go
package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var (
	execCommand = exec.Command
	osExit      = os.Exit
)

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	mode := fs.String("mode", "ci", "ci or fix")
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		osExit(1)
	}

	files, err := goFiles()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		osExit(1)
	}
	failed := false
	for _, f := range files {
		var e error
		switch *mode {
		case "ci":
			e = checkFile(f)
		case "fix":
			e = fixFile(f)
		default:
			fmt.Fprintf(os.Stderr, "unknown mode %q\n", *mode)
			osExit(2)
		}
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			failed = true
		}
	}
	if failed && *mode == "ci" {
		osExit(1)
	}
}

func goFiles() ([]string, error) {
	if _, err := execCommand("git", "--version").Output(); err != nil {
		return nil, errors.New("commentcheck requires git")
	}
	cmd := execCommand("git", "ls-files", "--", "*.go")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	var files []string
	for scanner.Scan() {
		files = append(files, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func checkFile(path string) error {
	rel := filepath.ToSlash(path)
	expected := "// " + rel

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("%s: %v", path, err)
	}

	groups := file.Comments
	if len(groups) == 0 {
		return fmt.Errorf("%s: missing file comment", path)
	}

	idx := 0
	if strings.HasPrefix(groups[0].List[0].Text, "//go:build") {
		for _, c := range groups[0].List {
			if !strings.HasPrefix(c.Text, "//go:build") {
				return fmt.Errorf("%s: invalid line %q in build tags", path, c.Text)
			}
		}
		idx = 1
	}

	if idx >= len(groups) {
		return fmt.Errorf("%s: missing file comment", path)
	}

	pathGroup := groups[idx]
	pos := fset.Position(pathGroup.Pos())
	if pathGroup.List[0].Text != expected {
		return fmt.Errorf("%s: first comment must be %q", path, expected)
	}
	if idx == 0 && pos.Line != 1 {
		return fmt.Errorf("%s: first comment must be %q", path, expected)
	}
	if idx > 0 && pos.Line < 2 {
		return fmt.Errorf("%s: first non-build comment must be %q", path, expected)
	}
	if len(pathGroup.List) > 1 || len(groups) > idx+1 {
		return fmt.Errorf("%s: found additional comments", path)
	}
	return nil
}

func fixFile(path string) error {
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
	comment := "// " + filepath.ToSlash(rel)
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
