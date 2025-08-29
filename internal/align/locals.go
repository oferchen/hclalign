// internal/align/locals.go
package align

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type localsStrategy struct{}

func (localsStrategy) Name() string { return "locals" }

func (localsStrategy) Align(_ *hclwrite.Block, _ *Options) error {
	return nil
}

func init() { Register(localsStrategy{}) }
