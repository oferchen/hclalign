package fs

import (
	"bytes"
	"context"
	iofs "io/fs"
	"os"
	"path/filepath"
)

// ApplyHints returns data adjusted to the desired newline style and BOM.
// The input data is assumed to use \n newlines and contain no BOM.
// If newline is nil or empty, \n is used. If bom is non-empty it is
// prepended to the result.
func ApplyHints(data, newline, bom []byte) []byte {
	out := make([]byte, len(data))
	copy(out, data)
	out = bytes.ReplaceAll(out, []byte("\r\n"), []byte("\n"))
	if len(newline) > 0 && !bytes.Equal(newline, []byte("\n")) {
		out = bytes.ReplaceAll(out, []byte("\n"), newline)
	}
	if len(bom) > 0 {
		out = append(append([]byte{}, bom...), out...)
	}
	return out
}

// WriteFile writes data to path atomically while preserving permissions.
// It writes to a temporary file in the same directory, syncs file and
// directory descriptors, and atomically renames it over the destination.
// The data is adjusted using the provided newline and BOM hints before
// writing.
func WriteFile(ctx context.Context, path string, data []byte, perm iofs.FileMode, newline, bom []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "hclalign-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	content := ApplyHints(data, newline, bom)
	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	dirf, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func() { _ = dirf.Close() }()
	return dirf.Sync()
}
