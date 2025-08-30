// internal/align/dynamic.go
package align

import "github.com/hashicorp/hcl/v2/hclwrite"

type dynamicStrategy struct{}

func (dynamicStrategy) Name() string { return "dynamic" }

func (dynamicStrategy) Align(block *hclwrite.Block, opts *Options) error {
	return nil
}

func init() { Register(dynamicStrategy{}) }
