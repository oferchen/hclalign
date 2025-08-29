// hack/stripcomments/main_test.go
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStrip(t *testing.T) {
	dir := t.TempDir()
	src := []byte(`// sample.go
package main

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
	expected := "// sample.go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hi\")\n}\n\n"
	if string(out) != expected {
		t.Fatalf("unexpected output:\n%s", out)
	}
}
