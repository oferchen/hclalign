// hack/stripcomments/main_test.go — SPDX-License-Identifier: Apache-2.0
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStrip(t *testing.T) {
	dir := t.TempDir()
	src := []byte(`package main

import "fmt"

// extra comment
func main() { // inline
    fmt.Println("hi") // trailing
}
`)
	path := filepath.Join(dir, "sample.go")
	if err := os.WriteFile(path, src, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := strip(dir); err != nil {
		t.Fatalf("strip: %v", err)
	}
	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("wd: %v", err)
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatalf("rel: %v", err)
	}
	expected := "// " + filepath.ToSlash(rel) + " — SPDX-License-Identifier: Apache-2.0\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hi\")\n}\n\n"
	if string(out) != expected {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestStripWithBuildTag(t *testing.T) {
	dir := t.TempDir()
	src := []byte("//go:build windows\n\npackage main\n\n// extra comment\nfunc main() {}\n")
	path := filepath.Join(dir, "sample.go")
	if err := os.WriteFile(path, src, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := strip(dir); err != nil {
		t.Fatalf("strip: %v", err)
	}
	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("wd: %v", err)
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatalf("rel: %v", err)
	}
	outStr := strings.ReplaceAll(string(out), "\t", "")
	need := "//go:build windows\n\n// " + filepath.ToSlash(rel) + " — SPDX-License-Identifier: Apache-2.0\npackage main\n\nfunc main(){}\n"
	if !strings.Contains(outStr, need) {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestProcessFilePreservesMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "perm.go")
	if err := os.WriteFile(path, []byte("package main\n"), 0755); err != nil {
		t.Fatalf("write file: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if err := processFile(path); err != nil {
		t.Fatalf("processFile: %v", err)
	}
	info2, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info2.Mode() != info.Mode() {
		t.Fatalf("mode changed: got %v, want %v", info2.Mode(), info.Mode())
	}
}
