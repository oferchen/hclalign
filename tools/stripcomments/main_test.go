// filename: tools/stripcomments/main_test.go
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcess(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "x.go")
	src := strings.Join([]string{
		"//go:build tag",
		"",
		"package main",
		"",
		"// pkg comment",
		"import \"fmt\"",
		"",
		"// main comment",
		"func main() { fmt.Println(\"hi\") } // inline",
	}, "\n") + "\n"
	if err := os.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := process(dir, file); err != nil {
		t.Fatalf("process: %v", err)
	}
	out, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	expected := strings.Join([]string{
		"//go:build tag",
		"",
		"// filename: x.go",
		"package main",
		"",
		"import \"fmt\"",
		"",
		"func main() { fmt.Println(\"hi\") }",
		"",
	}, "\n")
	if string(out) != expected {
		t.Fatalf("got %q want %q", out, expected)
	}
}

func TestMainRepoRoot(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "x.go")
	src := strings.Join([]string{
		"//go:build tag",
		"",
		"package main",
		"",
		"// pkg comment",
		"import \"fmt\"",
		"",
		"// main comment",
		"func main() { fmt.Println(\"hi\") } // inline",
	}, "\n") + "\n"
	if err := os.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	toolDir := filepath.Join(dir, "tools", "stripcomments")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	ignoreFile := filepath.Join(toolDir, "ignore.go")
	ignoreSrc := "package ignore\n// keep\n"
	if err := os.WriteFile(ignoreFile, []byte(ignoreSrc), 0o644); err != nil {
		t.Fatalf("write ignore: %v", err)
	}
	old := os.Args
	defer func() { os.Args = old }()
	os.Args = []string{"stripcomments", "--repo-root", dir}
	main()
	out, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	expected := strings.Join([]string{
		"//go:build tag",
		"",
		"// filename: x.go",
		"package main",
		"",
		"import \"fmt\"",
		"",
		"func main() { fmt.Println(\"hi\") }",
		"",
	}, "\n")
	if string(out) != expected {
		t.Fatalf("got %q want %q", out, expected)
	}
	ignored, err := os.ReadFile(ignoreFile)
	if err != nil {
		t.Fatalf("read ignore: %v", err)
	}
	if string(ignored) != ignoreSrc {
		t.Fatalf("ignored file changed: got %q want %q", ignored, ignoreSrc)
	}
}
