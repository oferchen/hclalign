//go:build windows

// internal/fs/atomic_windows_test.go
package fs

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestWriteFileAtomicChownIgnored(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")

	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatalf("prewrite: %v", err)
	}

	if err := os.Chown(path, 0, 0); !errors.Is(err, syscall.EWINDOWS) && !os.IsPermission(err) {
		t.Fatalf("unexpected chown error: %v", err)
	}

	if err := WriteFileAtomic(context.Background(), path, []byte("new"), 0o644, Hints{}); err != nil {
		t.Fatalf("WriteFileAtomic: %v", err)
	}
}
