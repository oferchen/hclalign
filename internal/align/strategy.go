package align

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/oferchen/hclalign/config"
)

// Strategy defines an alignment strategy that can transform an HCL file.
type Strategy func(*hclwrite.File, *config.Config) error

var strategies []Strategy

// Register adds s to the strategy registry. Typically called from init().
func Register(s Strategy) {
	strategies = append(strategies, s)
}

// Apply runs all registered strategies sequentially against the given file.
func Apply(file *hclwrite.File, cfg *config.Config) error {
	for _, s := range strategies {
		if err := s(file, cfg); err != nil {
			return err
		}
	}
	return nil
}
