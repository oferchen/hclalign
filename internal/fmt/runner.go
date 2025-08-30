// filename: internal/fmt/runner.go
package terraformfmt

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"unicode/utf8"

	"github.com/oferchen/hclalign/formatter"
	internalfs "github.com/oferchen/hclalign/internal/fs"
)

func Run(ctx context.Context, filename string, src []byte) ([]byte, error) {
	if _, err := exec.LookPath("terraform"); err != nil {
		return formatter.Format(src, filename)
	}
	hints := internalfs.DetectHintsFromBytes(src)
	src = internalfs.PrepareForParse(src, hints)
	if len(src) > 0 && !utf8.Valid(src) {
		return nil, fmt.Errorf("input is not valid UTF-8")
	}
	dir := filepath.Dir(filename)
	tmp, err := os.CreateTemp(dir, "hclalign-*.tf")
	if err != nil {
		return nil, err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(src); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return nil, err
	}
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "terraform", "fmt", "-no-color", "-list=false", "-write=true", "-diff=false", tmpName)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		os.Remove(tmpName)
		return nil, fmt.Errorf("terraform fmt failed: %v: %s", err, stderr.String())
	}
	formatted, err := os.ReadFile(tmpName)
	os.Remove(tmpName)
	if err != nil {
		return nil, err
	}
	formatted = internalfs.ApplyHints(formatted, hints)
	return formatted, nil
}
