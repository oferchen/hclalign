// internal/engine/reader.go
package engine

import (
	"context"
	"io"

	"github.com/oferchen/hclalign/config"
)

func ProcessReader(ctx context.Context, r io.Reader, w io.Writer, cfg *config.Config) (bool, error) {
	return processReader(ctx, r, w, cfg)
}
