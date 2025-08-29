// internal/align/terraform.go
package align

import (
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type terraformStrategy struct{}

func (terraformStrategy) Name() string { return "terraform" }

// Align orders attributes and nested blocks within a terraform block. Attributes
// are sorted alphabetically. Nested blocks are sorted by type and labels while
// preserving the ordering of their contents. If a required_providers block is
// present, the provider entries inside it are also sorted alphabetically.
func (terraformStrategy) Align(block *hclwrite.Block, _ *Options) error {
	body := block.Body()

	// Sort provider entries within required_providers blocks
	for _, nb := range body.Blocks() {
		if nb.Type() == "required_providers" {
			attrs := nb.Body().Attributes()
			names := make([]string, 0, len(attrs))
			for name := range attrs {
				names = append(names, name)
			}
			sort.Strings(names)
			if err := reorderBlock(nb, names); err != nil {
				return err
			}
		}
	}

	// Order top-level blocks by type then labels
	blocks := body.Blocks()
	sort.SliceStable(blocks, func(i, j int) bool {
		bi, bj := blocks[i], blocks[j]
		if bi.Type() != bj.Type() {
			return bi.Type() < bj.Type()
		}
		return strings.Join(bi.Labels(), "\x00") < strings.Join(bj.Labels(), "\x00")
	})
	for _, b := range body.Blocks() {
		body.RemoveBlock(b)
	}
	for _, b := range blocks {
		body.AppendBlock(b)
	}

	// Gather and order attributes
	attrs := body.Attributes()
	names := make([]string, 0, len(attrs))
	for name := range attrs {
		names = append(names, name)
	}
	sort.Strings(names)

	return reorderBlock(block, names)
}

func init() { Register(terraformStrategy{}) }
