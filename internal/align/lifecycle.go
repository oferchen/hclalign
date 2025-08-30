// internal/align/lifecycle.go
package align

import "github.com/hashicorp/hcl/v2/hclwrite"

type lifecycleStrategy struct{}

func (lifecycleStrategy) Name() string { return "lifecycle" }

func (lifecycleStrategy) Align(block *hclwrite.Block, opts *Options) error {
	return nil
}

func init() { Register(lifecycleStrategy{}) }
