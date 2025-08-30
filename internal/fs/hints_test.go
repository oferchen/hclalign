// filename: internal/fs/hints_test.go
package fs

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectHintsFromBytes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input []byte
		want  Hints
	}{
		{
			name:  "lf",
			input: []byte("one\ntwo\n"),
			want:  Hints{Newline: "\n"},
		},
		{
			name:  "crlf",
			input: []byte("one\r\ntwo\r\n"),
			want:  Hints{Newline: "\r\n"},
		},
		{
			name:  "bom_lf",
			input: append(append([]byte{}, utf8BOM...), []byte("one\ntwo\n")...),
			want:  Hints{HasBOM: true, Newline: "\n"},
		},
		{
			name:  "bom_crlf",
			input: append(append([]byte{}, utf8BOM...), []byte("one\r\ntwo\r\n")...),
			want:  Hints{HasBOM: true, Newline: "\r\n"},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := DetectHintsFromBytes(tc.input)
			if got.HasBOM != tc.want.HasBOM || got.Newline != tc.want.Newline {
				t.Fatalf("DetectHintsFromBytes returned %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestReadFileWithHints(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		content  []byte
		mode     os.FileMode
		wantData []byte
		wantHint Hints
	}{
		{
			name:     "bom_crlf",
			content:  append(append([]byte{}, utf8BOM...), []byte("one\r\ntwo\r\n")...),
			mode:     0o600,
			wantData: []byte("one\r\ntwo\r\n"),
			wantHint: Hints{HasBOM: true, Newline: "\r\n"},
		},
		{
			name:     "plain",
			content:  []byte("one\ntwo\n"),
			mode:     0o644,
			wantData: []byte("one\ntwo\n"),
			wantHint: Hints{Newline: "\n"},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "in.txt")
			if err := os.WriteFile(path, tc.content, tc.mode); err != nil {
				t.Fatalf("write file: %v", err)
			}
			data, perm, hint, err := ReadFileWithHints(context.Background(), path)
			if err != nil {
				t.Fatalf("ReadFileWithHints: %v", err)
			}
			if !bytes.Equal(data, tc.wantData) {
				t.Fatalf("data mismatch: %q != %q", data, tc.wantData)
			}
			if perm != tc.mode {
				t.Fatalf("mode mismatch: got %v want %v", perm, tc.mode)
			}
			if hint.HasBOM != tc.wantHint.HasBOM || hint.Newline != tc.wantHint.Newline {
				t.Fatalf("hints mismatch: got %+v want %+v", hint, tc.wantHint)
			}
		})
	}
}

func TestReadFileWithHintsCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, _, err := ReadFileWithHints(ctx, "irrelevant")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ReadFileWithHints error = %v; want %v", err, context.Canceled)
	}
}

func TestHintsBOM(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		hints Hints
		want  []byte
	}{
		{
			name:  "with_bom",
			hints: Hints{HasBOM: true},
			want:  utf8BOM,
		},
		{
			name:  "without_bom",
			hints: Hints{HasBOM: false},
			want:  nil,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.hints.BOM()
			if tc.hints.HasBOM {
				if !bytes.Equal(got, utf8BOM) {
					t.Fatalf("BOM returned %v, want %v", got, utf8BOM)
				}
			} else {
				if len(got) != 0 {
					t.Fatalf("BOM returned %v, want empty", got)
				}
			}
		})
	}
}
