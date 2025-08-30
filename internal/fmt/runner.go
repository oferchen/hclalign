// internal/fmt/runner.go
package terraformfmt

import (
        "context"
        "os/exec"

        internalfs "github.com/oferchen/hclalign/internal/fs"
        "github.com/oferchen/hclalign/formatter"
)

func Run(ctx context.Context, src []byte) ([]byte, internalfs.Hints, error) {
        if err := ctx.Err(); err != nil {
                return nil, internalfs.Hints{}, err
        }
        if _, err := exec.LookPath("terraform"); err == nil {
                return formatBinary(ctx, src)
        }
        formatted, hints, err := formatter.Format(src, "")
        if err != nil {
                return nil, internalfs.Hints{}, err
        }
        return formatted, hints, nil
}
