// internal/fs/stdio.go
package fs

import "io"

func WriteAllWithHints(w io.Writer, data []byte, hints Hints) error {
	content := ApplyHints(data, hints)
	_, err := w.Write(content)
	return err
}
