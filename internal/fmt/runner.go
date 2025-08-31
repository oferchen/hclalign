// internal/fmt/runner.go
package terraformfmt

import (
	"context"
	"os/exec"

	"github.com/oferchen/hclalign/formatter"
	internalfs "github.com/oferchen/hclalign/internal/fs"
)

func Run(ctx context.Context, src []byte) ([]byte, internalfs.Hints, error) {
	if err := ctx.Err(); err != nil {
		return nil, internalfs.Hints{}, err
	}
	if _, err := exec.LookPath("terraform"); err == nil {
		return formatBinary(ctx, src)
	}
	return formatter.Format(src, "")
}
