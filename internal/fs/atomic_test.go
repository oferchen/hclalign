package fs

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type setupFunc func(t *testing.T, dir, path string) any
type validateFunc func(t *testing.T, dir, path string, ctx any)

func TestWriteFileAtomic(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		data     []byte
		perm     os.FileMode
		hints    Hints
		setup    setupFunc
		validate validateFunc
	}{
		{
			name:  "bom and crlf",
			data:  []byte("one\ntwo\n"),
			perm:  0o644,
			hints: Hints{HasBOM: true, Newline: "\r\n"},
			validate: func(t *testing.T, dir, path string, _ any) {
				got, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read file: %v", err)
				}
				want := append([]byte{0xEF, 0xBB, 0xBF}, []byte("one\r\ntwo\r\n")...)
				if !bytes.Equal(got, want) {
					t.Fatalf("content mismatch: %q != %q", got, want)
				}
				info, err := os.Stat(path)
				if err != nil {
					t.Fatalf("stat: %v", err)
				}
				if info.Mode() != 0o644 {
					t.Fatalf("mode mismatch: got %v want %v", info.Mode(), os.FileMode(0o644))
				}
			},
		},
		{
			name: "same dir temp",
			data: []byte("x"),
			perm: 0o600,
			setup: func(t *testing.T, dir, path string) any {
				entries, err := os.ReadDir(os.TempDir())
				if err != nil {
					t.Fatalf("readdir: %v", err)
				}
				names := make(map[string]struct{}, len(entries))
				for _, e := range entries {
					names[e.Name()] = struct{}{}
				}
				return names
			},
			validate: func(t *testing.T, dir, path string, ctx any) {
				names := ctx.(map[string]struct{})
				entries, err := os.ReadDir(os.TempDir())
				if err != nil {
					t.Fatalf("readdir: %v", err)
				}
				for _, e := range entries {
					if strings.HasPrefix(e.Name(), "hclalign-") {
						if _, ok := names[e.Name()]; !ok {
							t.Fatalf("temp file created outside target dir: %s", e.Name())
						}
					}
				}
				dirEntries, err := os.ReadDir(dir)
				if err != nil {
					t.Fatalf("readdir target: %v", err)
				}
				if len(dirEntries) != 1 || dirEntries[0].Name() != filepath.Base(path) {
					t.Fatalf("unexpected files in target dir: %v", dirEntries)
				}
			},
		},
		{
			name: "permission retained",
			data: []byte("secret"),
			perm: 0o751,
			validate: func(t *testing.T, dir, path string, _ any) {
				info, err := os.Stat(path)
				if err != nil {
					t.Fatalf("stat: %v", err)
				}
				if info.Mode() != 0o751 {
					t.Fatalf("mode mismatch: got %v want %v", info.Mode(), os.FileMode(0o751))
				}
				got, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read file: %v", err)
				}
				if !bytes.Equal(got, []byte("secret")) {
					t.Fatalf("content mismatch: %q != %q", got, "secret")
				}
			},
		},
		{
			name: "rename semantics",
			data: []byte("new"),
			perm: 0o644,
			setup: func(t *testing.T, dir, path string) any {
				if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
					t.Fatalf("prewrite: %v", err)
				}
				return nil
			},
			validate: func(t *testing.T, dir, path string, _ any) {
				got, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read file: %v", err)
				}
				if !bytes.Equal(got, []byte("new")) {
					t.Fatalf("rename failed: %q != %q", got, "new")
				}
				entries, err := os.ReadDir(dir)
				if err != nil {
					t.Fatalf("readdir target: %v", err)
				}
				if len(entries) != 1 || entries[0].Name() != filepath.Base(path) {
					t.Fatalf("unexpected files in target dir: %v", entries)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "out.txt")
			var ctx any
			if tc.setup != nil {
				ctx = tc.setup(t, dir, path)
			}
			if err := WriteFileAtomic(path, tc.data, tc.perm, tc.hints); err != nil {
				t.Fatalf("WriteFileAtomic: %v", err)
			}
			if tc.validate != nil {
				tc.validate(t, dir, path, ctx)
			}
		})
	}
}
