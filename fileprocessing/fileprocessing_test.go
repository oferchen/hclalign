package fileprocessing

import (
	"bytes"
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
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

func TestProcessLogsFilesWhenVerbose(t *testing.T) {
	dir := t.TempDir()

	file1 := filepath.Join(dir, "a.tf")
	if err := os.WriteFile(file1, []byte("variable \"a\" {\n type = string\n}\n"), 0644); err != nil {
		t.Fatalf("write file1: %v", err)
	}
	file2 := filepath.Join(dir, "b.tf")
	if err := os.WriteFile(file2, []byte("variable \"b\" {\n type = string\n}\n"), 0644); err != nil {
		t.Fatalf("write file2: %v", err)
	}

	cfg := &config.Config{
		Target:      dir,
		Mode:        config.ModeCheck,
		Include:     config.DefaultInclude,
		Exclude:     config.DefaultExclude,
		Order:       config.DefaultOrder,
		Concurrency: 2,
		Verbose:     true,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	var buf bytes.Buffer
	old := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)

	if _, err := Process(context.Background(), cfg); err != nil {
		t.Fatalf("process: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "a.tf") || !strings.Contains(out, "b.tf") {
		t.Fatalf("expected log to contain file names, got: %q", out)
	}
}

func TestProcessContextCanceledNoLog(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.tf")
	if err := os.WriteFile(path, []byte("variable \"a\" {\n type = string\n}\n"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	cfg := &config.Config{
		Target:      dir,
		Mode:        config.ModeCheck,
		Include:     config.DefaultInclude,
		Exclude:     config.DefaultExclude,
		Order:       config.DefaultOrder,
		Concurrency: 1,
		Verbose:     true,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var buf bytes.Buffer
	old := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)

	if _, err := Process(ctx, cfg); err == nil {
		t.Fatalf("expected error due to canceled context")
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no logs, got %q", buf.String())
	}
}
