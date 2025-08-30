// filename: internal/fs/atomic.go
package fs

import (
	"bytes"
	"context"
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
	"syscall"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

type Hints struct {
	HasBOM  bool
	Newline string
}

func (h Hints) BOM() []byte {
	if h.HasBOM {
		return utf8BOM
	}
	return nil
}

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

func ReadAllWithHints(r io.Reader) ([]byte, Hints, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, Hints{}, err
	}
	h := DetectHintsFromBytes(raw)
	if h.HasBOM {
		raw = raw[len(utf8BOM):]
	}
	return raw, h, nil
}

func PrepareForParse(data []byte, hints Hints) []byte {
	if bom := hints.BOM(); len(bom) > 0 && bytes.HasPrefix(data, bom) {
		data = data[len(bom):]
	}
	return bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
}

func IsStdin(r io.Reader) bool {
	if f, ok := r.(*os.File); ok {
		return f.Fd() == os.Stdin.Fd()
	}
	return false
}

func IsStdout(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return f.Fd() == os.Stdout.Fd()
	}
	return false
}

func ReadFileWithHints(ctx context.Context, path string) (data []byte, perm iofs.FileMode, hints Hints, err error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, hints, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, 0, hints, err
	}
	perm = info.Mode()
	if err := ctx.Err(); err != nil {
		return nil, 0, hints, err
	}
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

type WriteOpts struct {
	Path  string
	Data  []byte
	Perm  iofs.FileMode
	Hints Hints
}

func WriteFileAtomic(ctx context.Context, opts WriteOpts) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path := opts.Path
	data := opts.Data
	perm := opts.Perm
	hints := opts.Hints
	dir := filepath.Dir(path)
	uid, gid := -1, -1
	if info, err := os.Stat(path); err == nil {
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			uid = int(stat.Uid)
			gid = int(stat.Gid)
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
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
	if uid != -1 || gid != -1 {
		if err := os.Chown(tmpName, uid, gid); err != nil {
			if !isErrWindows(err) && !os.IsPermission(err) {
				_ = tmp.Close()
				return err
			}
		}
	}
	if err := ctx.Err(); err != nil {
		_ = tmp.Close()
		return err
	}
	content := ApplyHints(data, hints)
	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := ctx.Err(); err != nil {
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
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	dirf, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func() { _ = dirf.Close() }()
	return dirf.Sync()
}
