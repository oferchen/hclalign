// cmd/commentcheck/main.go
package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var (
	execCommand = exec.Command
	lookPath    = exec.LookPath
	packageDirs = packageDirsFunc
	checkFile   = checkFileFunc
	osExit      = os.Exit
)

func main() {
	dirs, err := packageDirs()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		osExit(1)
	}
	var files []string
	for _, d := range dirs {
		entries, err := os.ReadDir(d)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			osExit(1)
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
		osExit(1)
	}
}

func checkFileFunc(path string) error {
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
	if strings.HasPrefix(first, "//go:build") {
		for {
			next, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("%s: unable to read after build tag", path)
			}
			line := strings.TrimRight(next, "\n")
			if line == "" {
				break
			}
			if !strings.HasPrefix(line, "//go:build") {
				return fmt.Errorf("%s: invalid line %q in build tags", path, line)
			}
		}
		pc, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("%s: unable to read path comment", path)
		}
		pc = strings.TrimRight(pc, "\n")
		if pc != expected {
			return fmt.Errorf("%s: path comment must be %q", path, expected)
		}
	} else {
		if first != expected {
			return fmt.Errorf("%s: first line must be %q", path, expected)
		}
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("%s: %v", path, err)
	}
	if len(file.Comments) == 0 {
		return fmt.Errorf("%s: missing file comment", path)
	}
	groups := file.Comments
	if strings.HasPrefix(first, "//go:build") {
		if len(groups) < 2 {
			return fmt.Errorf("%s: missing file comment", path)
		}
		buildGroup := groups[0]
		for _, c := range buildGroup.List {
			if !strings.HasPrefix(c.Text, "//go:build") {
				return fmt.Errorf("%s: unexpected comment %q in build tags", path, c.Text)
			}
		}
		pathGroup := groups[1]
		pos := fset.Position(pathGroup.Pos())
		if pathGroup.List[0].Text != expected || pos.Line < 2 {
			return fmt.Errorf("%s: first non-build comment must be %q", path, expected)
		}
		if len(groups) > 2 || len(pathGroup.List) > 1 {
			return fmt.Errorf("%s: found additional comments", path)
		}
	} else {
		firstGroup := groups[0]
		pos := fset.Position(firstGroup.Pos())
		if pos.Line != 1 || firstGroup.List[0].Text != expected {
			return fmt.Errorf("%s: first comment must be %q", path, expected)
		}
		if len(groups) > 1 || len(firstGroup.List) > 1 {
			return fmt.Errorf("%s: found additional comments", path)
		}
	}
	return nil
}

func packageDirsFunc() ([]string, error) {
	if _, err := lookPath("git"); err != nil {
		return nil, fmt.Errorf("commentcheck requires git: %w", err)
	}
	cmd := execCommand("git", "ls-files", "--", "*.go")
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.Error); ok && ee.Err == exec.ErrNotFound {
			return nil, errors.New("commentcheck requires git")
		}
		return nil, err
	}
	dirsSet := make(map[string]struct{})
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		dir := filepath.Dir(scanner.Text())
		dirsSet[dir] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	dirs := make([]string, 0, len(dirsSet))
	for d := range dirsSet {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	return dirs, nil
}
