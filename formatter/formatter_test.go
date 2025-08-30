// formatter/formatter_test.go
package formatter

import (
	"bytes"
	"testing"

	internalfs "github.com/oferchen/hclalign/internal/fs"
)

func TestFormatHints(t *testing.T) {
	bom := []byte{0xEF, 0xBB, 0xBF}
	tests := []struct {
		name      string
		input     []byte
		want      []byte
		wantHints internalfs.Hints
	}{
		{
			name:      "lf",
			input:     []byte("a=1\nb=2\n"),
			want:      []byte("a = 1\nb = 2\n"),
			wantHints: internalfs.Hints{Newline: "\n"},
		},
		{
			name:      "crlf_bom",
			input:     append(append([]byte{}, bom...), []byte("a=1\r\n")...),
			want:      append(append([]byte{}, bom...), []byte("a = 1\r\n")...),
			wantHints: internalfs.Hints{HasBOM: true, Newline: "\r\n"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, hints, err := Format(tc.input, "test.hcl")
			if err != nil {
				t.Fatalf("Format returned error: %v", err)
			}
			if hints != tc.wantHints {
				t.Fatalf("unexpected hints: got %+v want %+v", hints, tc.wantHints)
			}
			styled := internalfs.ApplyHints(got, hints)
			if !bytes.Equal(styled, tc.want) {
				t.Fatalf("unexpected output\nwant: %q\n got: %q", tc.want, styled)
			}
		})
	}
}

func TestFormatTrailingNewline(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		want      []byte
		wantHints internalfs.Hints
	}{
		{
			name:      "missing",
			input:     []byte("a=1"),
			want:      []byte("a = 1\n"),
			wantHints: internalfs.Hints{Newline: "\n"},
		},
		{
			name:      "extra",
			input:     []byte("a=1\n\n"),
			want:      []byte("a = 1\n"),
			wantHints: internalfs.Hints{Newline: "\n"},
		},
		{
			name:      "crlf_style",
			input:     []byte("a=1\r\nb=2"),
			want:      []byte("a = 1\r\nb = 2\r\n"),
			wantHints: internalfs.Hints{Newline: "\r\n"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, hints, err := Format(tc.input, "test.hcl")
			if err != nil {
				t.Fatalf("Format returned error: %v", err)
			}
			if hints != tc.wantHints {
				t.Fatalf("unexpected hints: got %+v want %+v", hints, tc.wantHints)
			}
			styled := internalfs.ApplyHints(got, hints)
			if !bytes.Equal(styled, tc.want) {
				t.Fatalf("unexpected output\nwant: %q\n got: %q", tc.want, styled)
			}
		})
	}
}

func TestFormatRejectsInvalidUTF8(t *testing.T) {
	invalid := []byte{0xff, 0xfe, 0xfd}
	if _, _, err := Format(invalid, "test.hcl"); err == nil {
		t.Fatalf("expected error for invalid UTF-8 input")
	}
}
