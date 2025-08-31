// internal/fs/stdio_test.go
package fs

import (
	"bytes"
	"testing"
)

func TestWriteAllWithHints(t *testing.T) {
	var buf bytes.Buffer
	err := WriteAllWithHints(&buf, []byte("a\n"), Hints{HasBOM: true, Newline: "\r\n"})
	if err != nil {
		t.Fatalf("WriteAllWithHints: %v", err)
	}
	want := append([]byte{0xEF, 0xBB, 0xBF}, []byte("a\r\n")...)
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("content mismatch: %q != %q", buf.Bytes(), want)
	}
}
