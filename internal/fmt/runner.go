// internal/fmt/runner.go
package terraformfmt

import (
	"context"
	"os/exec"
	"sync"

	"github.com/oferchen/hclalign/formatter"
	internalfs "github.com/oferchen/hclalign/internal/fs"
)

var terraformPath string
var terraformPathOnce sync.Once

func terraformBinary() string {
	terraformPathOnce.Do(func() {
		terraformPath, _ = exec.LookPath("terraform")
	})
	return terraformPath
}

func Run(ctx context.Context, src []byte) ([]byte, internalfs.Hints, error) {
	if err := ctx.Err(); err != nil {
		return nil, internalfs.Hints{}, err
	}
	if terraformBinary() != "" {
		return formatBinary(ctx, src)
	}
	return formatter.Format(src, "")
}
