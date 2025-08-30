// internal/fmt/runner.go
package terraformfmt

import (
	"context"

	"github.com/oferchen/hclalign/formatter"
)

func Run(ctx context.Context, src []byte) ([]byte, error) {
	return formatter.Format(src, "")
}
