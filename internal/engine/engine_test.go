// internal/engine/engine_test.go
package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/oferchen/hclalign/config"
	internalfs "github.com/oferchen/hclalign/internal/fs"
	"github.com/oferchen/hclalign/internal/hclalign"
	"github.com/stretchr/testify/require"
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
		Order:       config.CanonicalOrder,
		Concurrency: 1,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	if _, err := Process(context.Background(), cfg); err != nil {
		t.Fatalf("process: %v", err)
	}

	_, _, hints, err := internalfs.ReadFileWithHints(context.Background(), path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if !hints.HasBOM {
		t.Fatalf("bom not preserved")
	}
	if hints.Newline != "\r\n" {
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
		Order:       config.CanonicalOrder,
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

func TestProcessDiffDeterministicOrder(t *testing.T) {
	dir := t.TempDir()

	fileB := filepath.Join(dir, "b.tf")
	if err := os.WriteFile(fileB, []byte("variable \"b\" {\n  type = number\n  default = 1\n}\n"), 0644); err != nil {
		t.Fatalf("write b.tf: %v", err)
	}
	fileA := filepath.Join(dir, "a.tf")
	if err := os.WriteFile(fileA, []byte("variable \"a\" {\n  type = number\n  default = 1\n}\n"), 0644); err != nil {
		t.Fatalf("write a.tf: %v", err)
	}

	cfg := &config.Config{
		Target:      dir,
		Mode:        config.ModeDiff,
		Include:     config.DefaultInclude,
		Exclude:     config.DefaultExclude,
		Order:       config.CanonicalOrder,
		Concurrency: 2,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	changed, err := Process(context.Background(), cfg)
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if !changed {
		t.Fatalf("expected changes")
	}

	_ = w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	outStr := string(out)
	idxA := strings.Index(outStr, "a.tf")
	idxB := strings.Index(outStr, "b.tf")
	if idxA == -1 || idxB == -1 {
		t.Fatalf("diff output missing: %q", outStr)
	}
	if idxA > idxB {
		t.Fatalf("expected a.tf diff before b.tf diff: %q", outStr)
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
		Order:       config.CanonicalOrder,
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

func TestProcessSingleFileContextCanceled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.tf")

	var buf bytes.Buffer
	buf.WriteString("variable \"a\" {\n")
	for i := 2000; i >= 0; i-- {
		fmt.Fprintf(&buf, "  attr%04d = %d\n", i, i)
	}
	buf.WriteString("}\n")
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	cfg := &config.Config{Order: config.CanonicalOrder, Mode: config.ModeCheck, Concurrency: 1}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, _, err := processSingleFile(ctx, path, cfg)
		done <- err
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()

	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled error, got %v", err)
	}
=======
func TestProcessMissingTarget(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.tf")
	cfg := &config.Config{
		Target:      missing,
		Mode:        config.ModeCheck,
		Include:     config.DefaultInclude,
		Exclude:     config.DefaultExclude,
		Order:       config.CanonicalOrder,
		Concurrency: 1,
	}
	require.NoError(t, cfg.Validate())
	_, err := Process(context.Background(), cfg)
	require.EqualError(t, err, fmt.Sprintf("target %q does not exist", missing))
=======
func TestProcessSingleFileCanceledAfterParse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.tf")
	require.NoError(t, os.WriteFile(path, []byte("variable \"a\" {\n type = string\n}\n"), 0644))

	ctx, cancel := context.WithCancel(context.Background())
	testHookAfterParse = func() { cancel() }
	defer func() { testHookAfterParse = nil }()

	called := false
	reorderAttributes = func(f *hclwrite.File, order []string, strict bool) error {
		called = true
		return nil
	}
	defer func() { reorderAttributes = hclalign.ReorderAttributes }()

	changed, out, err := processSingleFile(ctx, path, &config.Config{Mode: config.ModeCheck, Order: config.CanonicalOrder})
	require.ErrorIs(t, err, context.Canceled)
	require.False(t, changed)
	require.Nil(t, out)
	require.False(t, called, "reorderAttributes should not be called on canceled context")
}

func TestProcessSingleFileCanceledAfterReorder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.tf")
	require.NoError(t, os.WriteFile(path, []byte("variable \"a\" {\n type = string\n}\n"), 0644))

	ctx, cancel := context.WithCancel(context.Background())
	testHookAfterReorder = func() { cancel() }
	defer func() { testHookAfterReorder = nil }()

	called := false
	reorderAttributes = func(f *hclwrite.File, order []string, strict bool) error {
		called = true
		return nil
	}
	defer func() { reorderAttributes = hclalign.ReorderAttributes }()

	changed, out, err := processSingleFile(ctx, path, &config.Config{Mode: config.ModeCheck, Order: config.CanonicalOrder})
	require.ErrorIs(t, err, context.Canceled)
	require.False(t, changed)
	require.Nil(t, out)
	require.True(t, called, "reorderAttributes should be called before cancellation")


}

func TestProcessPropagatesFileError(t *testing.T) {
	dir := t.TempDir()

	good := filepath.Join(dir, "good.tf")
	if err := os.WriteFile(good, []byte("variable \"a\" {\n type = string\n}\n"), 0644); err != nil {
		t.Fatalf("write good file: %v", err)
	}
	bad := filepath.Join(dir, "bad.tf")
	if err := os.WriteFile(bad, []byte("variable \"b\" {\n"), 0644); err != nil {
		t.Fatalf("write bad file: %v", err)
	}

	cfg := &config.Config{
		Target:      dir,
		Mode:        config.ModeCheck,
		Include:     config.DefaultInclude,
		Exclude:     config.DefaultExclude,
		Order:       config.CanonicalOrder,
		Concurrency: 2,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	if _, err := Process(context.Background(), cfg); err == nil {
		t.Fatalf("expected error due to invalid file")
	}
}


func TestProcessHaltsAfterMalformedFile(t *testing.T) {
	dir := t.TempDir()

	fileA := filepath.Join(dir, "a.tf")
	contentA := "variable \"a\" {\n  default = 1\n  type = number\n}\n"
	if err := os.WriteFile(fileA, []byte(contentA), 0644); err != nil {
		t.Fatalf("write a.tf: %v", err)
	}
	fileB := filepath.Join(dir, "b.tf")
	if err := os.WriteFile(fileB, []byte("variable \"b\" {\n"), 0644); err != nil {
		t.Fatalf("write b.tf: %v", err)
	}
	fileC := filepath.Join(dir, "c.tf")
	contentC := "variable \"c\" {\n  type = number\n  default = 1\n}\n"
	if err := os.WriteFile(fileC, []byte(contentC), 0644); err != nil {
		t.Fatalf("write c.tf: %v", err)
	}

	cfg := &config.Config{
		Target:      dir,
		Mode:        config.ModeWrite,
		Include:     config.DefaultInclude,
		Exclude:     config.DefaultExclude,
		Order:       config.CanonicalOrder,
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

	if _, err := Process(context.Background(), cfg); err == nil {
		t.Fatalf("expected error due to invalid file")
	}

	logs := buf.String()
	if !strings.Contains(logs, "processed file: "+fileA) {
		t.Fatalf("expected log for a.tf, got %q", logs)
	}
	if strings.Contains(logs, "processed file: "+fileC) {
		t.Fatalf("did not expect log for c.tf, got %q", logs)
	}

	data, err := os.ReadFile(fileC)
	if err != nil {
		t.Fatalf("read c.tf: %v", err)
	}
	if string(data) != contentC {
		t.Fatalf("expected c.tf to remain unchanged, got %q", string(data))
	}
=======
func TestProcessMissingTarget(t *testing.T) {
	t.Run("nonexistent path", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "missing.tf")

		cfg := &config.Config{
			Target:      target,
			Mode:        config.ModeCheck,
			Include:     config.DefaultInclude,
			Exclude:     config.DefaultExclude,
			Order:       config.CanonicalOrder,
			Concurrency: 1,
		}
		require.NoError(t, cfg.Validate())
		_, err := Process(context.Background(), cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("target %q does not exist", target))
	})

	t.Run("broken symlink", func(t *testing.T) {
		dir := t.TempDir()
		link := filepath.Join(dir, "broken.tf")
		if err := os.Symlink(filepath.Join(dir, "missing.tf"), link); err != nil {
			t.Fatalf("symlink: %v", err)
		}

		cfg := &config.Config{
			Target:      link,
			Mode:        config.ModeCheck,
			Include:     config.DefaultInclude,
			Exclude:     config.DefaultExclude,
			Order:       config.CanonicalOrder,
			Concurrency: 1,
		}
		require.NoError(t, cfg.Validate())
		_, err := Process(context.Background(), cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("target %q does not exist", link))
	})

}

func TestProcessSymlinkedDirTargetFollowSymlinks(t *testing.T) {
	dir := t.TempDir()
	realDir := filepath.Join(dir, "real")
	if err := os.Mkdir(realDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(realDir, "a.tf")
	if err := os.WriteFile(path, []byte("variable \"a\" {\n  default = 1\n  type = number\n}\n"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(realDir, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	cfg := &config.Config{
		Target:         link,
		Mode:           config.ModeCheck,
		Include:        config.DefaultInclude,
		Exclude:        config.DefaultExclude,
		Order:          config.CanonicalOrder,
		Concurrency:    1,
		FollowSymlinks: true,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	changed, err := Process(context.Background(), cfg)
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if !changed {
		t.Fatalf("expected changes when following symlinked directory")
	}
}

func TestProcessSymlinkedDirTargetNoFollow(t *testing.T) {
	dir := t.TempDir()
	realDir := filepath.Join(dir, "real")
	if err := os.Mkdir(realDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(realDir, "a.tf")
	if err := os.WriteFile(path, []byte("variable \"a\" {\n  default = 1\n  type = number\n}\n"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(realDir, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	cfg := &config.Config{
		Target:      link,
		Mode:        config.ModeCheck,
		Include:     config.DefaultInclude,
		Exclude:     config.DefaultExclude,
		Order:       config.CanonicalOrder,
		Concurrency: 1,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	changed, err := Process(context.Background(), cfg)
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if changed {
		t.Fatalf("did not expect changes when not following symlinked directory")
	}
}

func TestProcessSymlinkedFileTargetFollowSymlinks(t *testing.T) {
	dir := t.TempDir()
	realFile := filepath.Join(dir, "a.tf")
	if err := os.WriteFile(realFile, []byte("variable \"a\" {\n  default = 1\n  type = number\n}\n"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	link := filepath.Join(dir, "link.tf")
	if err := os.Symlink(realFile, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	cfg := &config.Config{
		Target:         link,
		Mode:           config.ModeCheck,
		Include:        config.DefaultInclude,
		Exclude:        config.DefaultExclude,
		Order:          config.CanonicalOrder,
		Concurrency:    1,
		FollowSymlinks: true,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	changed, err := Process(context.Background(), cfg)
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if !changed {
		t.Fatalf("expected changes when processing symlinked file with FollowSymlinks")
	}
}

func TestProcessSymlinkedFileTargetNoFollow(t *testing.T) {
	dir := t.TempDir()
	realFile := filepath.Join(dir, "a.tf")
	if err := os.WriteFile(realFile, []byte("variable \"a\" {\n  default = 1\n  type = number\n}\n"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	link := filepath.Join(dir, "link.tf")
	if err := os.Symlink(realFile, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	cfg := &config.Config{
		Target:      link,
		Mode:        config.ModeCheck,
		Include:     config.DefaultInclude,
		Exclude:     config.DefaultExclude,
		Order:       config.CanonicalOrder,
		Concurrency: 1,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	changed, err := Process(context.Background(), cfg)
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if !changed {
		t.Fatalf("expected changes when processing symlinked file without FollowSymlinks")
	}
}

func TestProcessReaderPreservesNewlineAndBOM(t *testing.T) {
	bom := []byte{0xEF, 0xBB, 0xBF}
	input := string(bom) + "variable \"a\" {\r\n  default = 1\r\n  type = number\r\n}\r\n"

	cfg := &config.Config{Mode: config.ModeWrite, Stdout: true, Order: config.CanonicalOrder}


	var buf bytes.Buffer
	if _, err := processReader(context.Background(), bytes.NewReader([]byte(input)), &buf, cfg); err != nil {
		t.Fatalf("processReader: %v", err)
	}
	out := buf.Bytes()
=======
	var out bytes.Buffer
	if _, err := processReader(context.Background(), bytes.NewReader([]byte(input)), &out, cfg); err != nil {
		t.Fatalf("processReader: %v", err)
	}
	data := out.Bytes()


	hints := internalfs.DetectHintsFromBytes(data)
	if !hints.HasBOM {
		t.Fatalf("bom not preserved")
	}
	if hints.Newline != "\r\n" {
		t.Fatalf("LF line ending found")
	}

	expectedFile, diags := hclwrite.ParseConfig([]byte("variable \"a\" {\n  default = 1\n  type = number\n}\n"), "stdin", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatalf("parse expected: %v", diags)
	}
	require.NoError(t, hclalign.ReorderAttributes(expectedFile, config.CanonicalOrder, false))
	expected := internalfs.ApplyHints(expectedFile.Bytes(), internalfs.Hints{HasBOM: true, Newline: "\r\n"})
	if string(data) != string(expected) {
		t.Fatalf("unexpected output: got %q, want %q", data, expected)
	}
}

func TestProcessReaderDiffPreservesNewline(t *testing.T) {
	bom := []byte{0xEF, 0xBB, 0xBF}
	input := string(bom) + "variable \"a\" {\r\n  default = 1\r\n  type = number\r\n}\r\n"

	cfg := &config.Config{Mode: config.ModeDiff, Order: config.CanonicalOrder}


	var buf bytes.Buffer
	changed, err := processReader(context.Background(), bytes.NewReader([]byte(input)), &buf, cfg)
=======
	var out bytes.Buffer
	changed, err := processReader(context.Background(), bytes.NewReader([]byte(input)), &out, cfg)

	if err != nil {
		t.Fatalf("processReader: %v", err)
	}
	if !changed {
		t.Fatalf("expected changes")
	}

	out := buf.Bytes()
	hints := internalfs.DetectHintsFromBytes(out)
=======
	hints := internalfs.DetectHintsFromBytes(out.Bytes())

	if hints.Newline != "\r\n" {
		t.Fatalf("expected CRLF in diff output")
	}
}

func TestProcessStopsAfterMalformedFile(t *testing.T) {
	dir := t.TempDir()

	good := filepath.Join(dir, "good.tf")
	if err := os.WriteFile(good, []byte("variable \"a\" {\n  default = 1\n  type = number\n}\n"), 0644); err != nil {
		t.Fatalf("write good file: %v", err)
	}
	bad := filepath.Join(dir, "bad.tf")
	if err := os.WriteFile(bad, []byte("variable \"b\" {\n  default = 1\n"), 0644); err != nil {
		t.Fatalf("write bad file: %v", err)
	}

	cfg := &config.Config{
		Target:      dir,
		Mode:        config.ModeWrite,
		Include:     config.DefaultInclude,
		Exclude:     config.DefaultExclude,
		Order:       config.CanonicalOrder,
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

	changed, procErr := Process(context.Background(), cfg)
	if procErr == nil {
		t.Fatalf("expected error")
	}
	if changed {
		t.Fatalf("did not expect changes")
	}

	out := buf.String()
	if strings.Contains(out, "processed file: "+good) {
		t.Fatalf("did not expect log for good file, got %q", out)
	}
	if !strings.Contains(out, "error processing file "+bad) {
		t.Fatalf("expected error log for bad file, got %q", out)
	}

	data, err := os.ReadFile(good)
	if err != nil {
		t.Fatalf("read good file: %v", err)
	}
	expected := []byte("variable \"a\" {\n  default = 1\n  type = number\n}\n")
	if string(data) != string(expected) {
		t.Fatalf("good file unexpectedly processed: got %q, want %q", data, expected)
	}
	if !strings.Contains(procErr.Error(), "bad.tf") {
		t.Fatalf("error does not mention bad file: %v", procErr)
	}
}

func TestProcessSkipsDefaultExcludedDirs(t *testing.T) {
	dir := t.TempDir()

	validFile := filepath.Join(dir, "main.tf")
	if err := os.WriteFile(validFile, []byte("variable \"a\" {}\n"), 0644); err != nil {
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
		Order:       config.CanonicalOrder,
		Concurrency: 1,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	if _, err := Process(context.Background(), cfg); err != nil {
		t.Fatalf("process: %v", err)
	}
}

func TestProcessStdoutError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.tf")
	if err := os.WriteFile(path, []byte("variable \"a\" {}\n"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	cfg := &config.Config{
		Mode:        config.ModeCheck,
		Include:     config.DefaultInclude,
		Exclude:     config.DefaultExclude,
		Order:       config.CanonicalOrder,
		Stdout:      true,
		Concurrency: 1,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	_ = w.Close()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	if _, err := Process(context.Background(), cfg); err == nil {
		t.Fatalf("expected error")
	}
	_ = r.Close()
}

func TestProcessReaderStdoutError(t *testing.T) {
	cfg := &config.Config{Mode: config.ModeWrite, Stdout: true, Order: config.CanonicalOrder}

	r, w := io.Pipe()
	_ = w.Close()

	input := "variable \"a\" {}\n"
=======

	input := "variable \"a\" {}\n"
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	_ = w.Close()

	if _, err := processReader(context.Background(), bytes.NewReader([]byte(input)), w, cfg); err == nil {
		t.Fatalf("expected error")
	}
	_ = r.Close()
}

func TestProcessStrictOrder(t *testing.T) {
	t.Run("unknown attribute", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "a.tf")
		err := os.WriteFile(path, []byte("variable \"a\" {\n  foo = 1\n}\n"), 0644)
		require.NoError(t, err)

		cfg := &config.Config{
			Target:      dir,
			Mode:        config.ModeCheck,
			Include:     config.DefaultInclude,
			Exclude:     config.DefaultExclude,
			Order:       config.CanonicalOrder,
			StrictOrder: true,
			Concurrency: 1,
		}
		require.NoError(t, cfg.Validate())

		_, err = Process(context.Background(), cfg)
		require.Error(t, err)
	})

	t.Run("valid", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "a.tf")
		content := "variable \"a\" {\n  description = \"d\"\n  type        = string\n  default     = \"v\"\n  sensitive   = false\n  nullable    = true\n}\n"
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)

		cfg := &config.Config{
			Target:      dir,
			Mode:        config.ModeCheck,
			Include:     config.DefaultInclude,
			Exclude:     config.DefaultExclude,
			Order:       config.CanonicalOrder,
			StrictOrder: true,
			Concurrency: 1,
		}
		require.NoError(t, cfg.Validate())

		_, err = Process(context.Background(), cfg)
		require.NoError(t, err)
	})
}
