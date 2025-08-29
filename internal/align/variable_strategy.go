package align

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/oferchen/hclalign/config"
	hclalignpkg "github.com/oferchen/hclalign/internal/hclalign"
)

// variableStrategy applies canonical variable attribute ordering.
func variableStrategy(file *hclwrite.File, cfg *config.Config) error {
	return hclalignpkg.ReorderAttributes(file, cfg.Order, cfg.StrictOrder)
}

func init() {
	Register(variableStrategy)
}
