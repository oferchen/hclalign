package engine

import (
	"context"
	"io"

	"github.com/oferchen/hclalign/config"
)

// ProcessReader processes HCL content from r and writes the result to w.
// It exposes the unexported processReader for external callers and tests.
func ProcessReader(ctx context.Context, r io.Reader, w io.Writer, cfg *config.Config) (bool, error) {
	return processReader(ctx, r, w, cfg)
}
