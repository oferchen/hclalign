package fileprocessing

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/oferchen/hclalign/config"
)

func TestProcessPreservesNewlineAndBOM(t *testing.T) {
	dir := t.TempDir()
	bom := []byte{0xEF, 0xBB, 0xBF}
	content := string(bom) + "variable \"a\" {\r\ntype = string\r\ndescription = \"desc\"\r\n}\r\n"
	path := filepath.Join(dir, "test.tf")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	cfg := &config.Config{
		Target:      dir,
		Mode:        config.ModeWrite,
		Include:     config.DefaultInclude,
		Exclude:     config.DefaultExclude,
		Order:       config.DefaultOrder,
		Concurrency: 1,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	if _, err := Process(context.Background(), cfg); err != nil {
		t.Fatalf("process: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if !bytes.HasPrefix(data, bom) {
		t.Fatalf("bom not preserved")
	}
	if bytes.Contains(bytes.ReplaceAll(data, []byte("\r\n"), []byte{}), []byte("\n")) {
		t.Fatalf("LF line ending found")
	}
}
