// internal/fmt/runner.go
package terraformfmt

import (
	"context"

	"github.com/oferchen/hclalign/formatter"
	internalfs "github.com/oferchen/hclalign/internal/fs"
)

func Run(ctx context.Context, src []byte) ([]byte, internalfs.Hints, error) {
	return formatter.Format(src, "")
}
