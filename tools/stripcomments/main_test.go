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

func TestMainNoArgs(t *testing.T) {
	old := os.Args
	defer func() { os.Args = old }()
	os.Args = []string{"stripcomments"}
	main()
}
