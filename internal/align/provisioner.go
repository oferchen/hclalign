// internal/align/provisioner.go
package align

import "github.com/hashicorp/hcl/v2/hclwrite"

type provisionerStrategy struct{}

func (provisionerStrategy) Name() string { return "provisioner" }

func (provisionerStrategy) Align(block *hclwrite.Block, opts *Options) error {
	return nil
}

func init() { Register(provisionerStrategy{}) }
