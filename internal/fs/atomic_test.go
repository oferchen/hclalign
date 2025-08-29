// internal/fs/atomic_test.go — SPDX-License-Identifier: Apache-2.0
package fs

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
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
			name: "ownership retained",
			data: []byte("new"),
			perm: 0o600,
			setup: func(t *testing.T, dir, path string) any {
				if err := os.WriteFile(path, []byte("old"), 0o600); err != nil {
					t.Fatalf("prewrite: %v", err)
				}
				if err := os.Chown(path, 1, 1); err != nil {
					t.Skipf("chown: %v", err)
				}
				info, err := os.Stat(path)
				if err != nil {
					t.Fatalf("stat: %v", err)
				}
				st, ok := info.Sys().(*syscall.Stat_t)
				if !ok {
					t.Skip("stat_t not available")
				}
				return [2]uint32{st.Uid, st.Gid}
			},
			validate: func(t *testing.T, dir, path string, ctx any) {
				want := ctx.([2]uint32)
				info, err := os.Stat(path)
				if err != nil {
					t.Fatalf("stat: %v", err)
				}
				st, ok := info.Sys().(*syscall.Stat_t)
				if !ok {
					t.Skip("stat_t not available")
				}
				if st.Uid != want[0] || st.Gid != want[1] {
					t.Fatalf("ownership mismatch: got (%d,%d) want (%d,%d)", st.Uid, st.Gid, want[0], want[1])
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
			if err := WriteFileAtomic(context.Background(), path, tc.data, tc.perm, tc.hints); err != nil {
				t.Fatalf("WriteFileAtomic: %v", err)
			}
			if tc.validate != nil {
				tc.validate(t, dir, path, ctx)
			}
		})
	}
}

func TestWriteFileAtomicContextCanceledBeforeRename(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatalf("prewrite: %v", err)
	}

	data := bytes.Repeat([]byte{'x'}, 100<<20)
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- WriteFileAtomic(ctx, path, data, 0o644, Hints{})
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()

	err := <-errCh
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("WriteFileAtomic error = %v; want %v", err, context.Canceled)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if !bytes.Equal(got, []byte("old")) {
		t.Fatalf("file modified: %q != %q", got, "old")
	}
}

func TestApplyHints(t *testing.T) {
	data := []byte("one\ntwo\n")

	got := ApplyHints(data, Hints{})
	if !bytes.Equal(got, data) {
		t.Fatalf("no-op mismatch: %q != %q", got, data)
	}

	got = ApplyHints(data, Hints{Newline: "\r\n"})
	want := []byte("one\r\ntwo\r\n")
	if !bytes.Equal(got, want) {
		t.Fatalf("newline mismatch: %q != %q", got, want)
	}

	got = ApplyHints(data, Hints{HasBOM: true})
	want = append([]byte{}, utf8BOM...)
	want = append(want, data...)
	if !bytes.Equal(got, want) {
		t.Fatalf("bom mismatch: %q != %q", got, want)
	}
}

func TestApplyHintsAlloc(t *testing.T) {
	data := []byte("one\ntwo\n")
	allocs := testing.AllocsPerRun(1000, func() {
		ApplyHints(data, Hints{})
	})
	if allocs != 0 {
		t.Fatalf("unexpected allocations: got %f want 0", allocs)
	}
}

func BenchmarkApplyHints(b *testing.B) {
	data := []byte("one\ntwo\n")
	hints := Hints{Newline: "\r\n"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ApplyHints(data, hints)
	}
}
