package fileprocessing

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/oferchen/hclalign/config"
	"github.com/oferchen/hclalign/hclprocessing"
	internalfs "github.com/oferchen/hclalign/internal/fs"
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

func TestProcessReaderPreservesNewlineAndBOM(t *testing.T) {
	bom := []byte{0xEF, 0xBB, 0xBF}
	input := string(bom) + "variable \"a\" {\r\n  default = 1\r\n  type = number\r\n}\r\n"

	cfg := &config.Config{Mode: config.ModeWrite, Stdout: true, Order: config.DefaultOrder}

	oldStdout := os.Stdout
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = wOut
	defer func() { os.Stdout = oldStdout }()

	if _, err := processReader(context.Background(), bytes.NewReader([]byte(input)), cfg); err != nil {
		t.Fatalf("processReader: %v", err)
	}
	_ = wOut.Close()
	out, err := io.ReadAll(rOut)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	if !bytes.HasPrefix(out, bom) {
		t.Fatalf("bom not preserved")
	}
	if bytes.Contains(bytes.ReplaceAll(out, []byte("\r\n"), []byte{}), []byte("\n")) {
		t.Fatalf("LF line ending found")
	}

	expectedFile, diags := hclwrite.ParseConfig([]byte("variable \"a\" {\n  default = 1\n  type = number\n}\n"), "stdin", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatalf("parse expected: %v", diags)
	}
	hclprocessing.ReorderAttributes(expectedFile, config.DefaultOrder, false)
	expected := internalfs.ApplyHints(expectedFile.Bytes(), internalfs.Hints{HasBOM: true, Newline: "\r\n"})
	if string(out) != string(expected) {
		t.Fatalf("unexpected output: got %q, want %q", out, expected)
	}
}

func TestProcessReaderDiffPreservesNewline(t *testing.T) {
	bom := []byte{0xEF, 0xBB, 0xBF}
	input := string(bom) + "variable \"a\" {\r\n  default = 1\r\n  type = number\r\n}\r\n"

	cfg := &config.Config{Mode: config.ModeDiff, Order: config.DefaultOrder}

	oldStdout := os.Stdout
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = wOut
	defer func() { os.Stdout = oldStdout }()

	changed, err := processReader(context.Background(), bytes.NewReader([]byte(input)), cfg)
	if err != nil {
		t.Fatalf("processReader: %v", err)
	}
	if !changed {
		t.Fatalf("expected changes")
	}
	_ = wOut.Close()
	out, err := io.ReadAll(rOut)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if !bytes.Contains(out, []byte("\r\n")) {
		t.Fatalf("expected CRLF in diff output")
	}
}

func TestProcessSkipsDefaultExcludedDirs(t *testing.T) {
	dir := t.TempDir()
	// valid file to ensure processing succeeds
	valid := filepath.Join(dir, "main.tf")
	if err := os.WriteFile(valid, []byte("variable \"a\" {}\n"), 0644); err != nil {
		t.Fatalf("write valid: %v", err)
	}
	excluded := []string{".git", ".terraform", "vendor", "node_modules"}
	for _, d := range excluded {
		path := filepath.Join(dir, d)
		if err := os.Mkdir(path, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
		bad := filepath.Join(path, "bad.tf")
		if err := os.WriteFile(bad, []byte("not hcl"), 0644); err != nil {
			t.Fatalf("write bad: %v", err)
		}
	}

	cfg := &config.Config{
		Target:      dir,
		Mode:        config.ModeCheck,
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
}
