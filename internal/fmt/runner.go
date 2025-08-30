// internal/fmt/runner.go
package terraformfmt

import (
	"context"
	"os/exec"

	"github.com/oferchen/hclalign/formatter"
)

func Run(ctx context.Context, src []byte) ([]byte, error) {
	if _, err := exec.LookPath("terraform"); err != nil {
		return formatter.Format(src, "")
	}
	return formatBinary(ctx, src)
}
