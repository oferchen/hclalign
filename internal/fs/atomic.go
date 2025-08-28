package fs

import (
	"bytes"
	iofs "io/fs"
	"os"
	"path/filepath"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// Hints describes how a file's content is structured. Newline indicates the
// newline sequence used in the file ("\n" or "\r\n"), and HasBOM specifies
// whether the file starts with a UTF-8 BOM.
type Hints struct {
	HasBOM  bool
	Newline string
}

// BOM returns the byte slice representing the BOM if one should be used.
func (h Hints) BOM() []byte {
	if h.HasBOM {
		return utf8BOM
	}
	return nil
}

// DetectHintsFromBytes inspects b and returns detected newline style and BOM
// presence. The returned hints default to "\n" newlines with no BOM.
func DetectHintsFromBytes(b []byte) Hints {
	h := Hints{Newline: "\n"}
	if len(b) >= len(utf8BOM) && bytes.Equal(b[:len(utf8BOM)], utf8BOM) {
		h.HasBOM = true
		b = b[len(utf8BOM):]
	}
	if bytes.Contains(b, []byte("\r\n")) {
		h.Newline = "\r\n"
	}
	return h
}

// ReadFileWithHints reads the file at path and returns its data without any
// leading BOM, its permissions, and detected newline/BOM hints.
func ReadFileWithHints(path string) (data []byte, perm iofs.FileMode, hints Hints, err error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, 0, hints, err
	}
	perm = info.Mode()
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, hints, err
	}
	hints = DetectHintsFromBytes(raw)
	if hints.HasBOM {
		raw = raw[len(utf8BOM):]
	}
	return raw, perm, hints, nil
}

// ApplyHints returns data adjusted to the desired newline style and BOM. The
// input data is assumed to use \n newlines and contain no BOM. If Newline is
// empty, \n is used. If HasBOM is true, a UTF-8 BOM is prepended to the result.
func ApplyHints(data []byte, hints Hints) []byte {
	out := data
	if hints.Newline == "\r\n" {
		out = bytes.ReplaceAll(out, []byte("\n"), []byte("\r\n"))
	}
	if hints.HasBOM {
		out = append(append([]byte{}, utf8BOM...), out...)
	}
	return out
}

// WriteFileAtomic writes data to path atomically while preserving permissions.
// It writes to a temporary file in the same directory, syncs file and
// directory descriptors, and atomically renames it over the destination. The
// data is adjusted using the provided hints before writing.
func WriteFileAtomic(path string, data []byte, perm iofs.FileMode, hints Hints) error {
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
	content := ApplyHints(data, hints)
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
