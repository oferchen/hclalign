package fs

import "io"

// WriteAllWithHints applies the provided hints to data and writes the result
// to w.
func WriteAllWithHints(w io.Writer, data []byte, hints Hints) error {
	content := ApplyHints(data, hints)
	_, err := w.Write(content)
	return err
}
