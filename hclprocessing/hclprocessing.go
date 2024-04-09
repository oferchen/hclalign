// hclprocessing.go
// Manages parsing and reordering of attributes within HCL files.

package hclprocessing

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"sort"
)

// ReorderAttributes reorders the attributes of the given HCL file according to the specified order.
func ReorderAttributes(file *hclwrite.File, order []string) {
	for _, block := range file.Body().Blocks() {
		reorderBlockAttributes(block, order)
	}
}

// reorderBlockAttributes reorders attributes within a single block based on the specified order.
func reorderBlockAttributes(block *hclwrite.Block, order []string) {
	originalAttributes := block.Body().Attributes()
	sortedAttributes := sortAttributes(originalAttributes, order)

	// Remove all attributes to re-add them in the correct order.
	for name := range originalAttributes {
		block.Body().RemoveAttribute(name)
	}

	// Re-add attributes in the specified order.
	for _, name := range sortedAttributes {
		if attr, exists := originalAttributes[name]; exists {
			block.Body().SetAttributeRaw(name, attr.Expr().BuildTokens(nil))
		}
	}
}

// sortAttributes sorts the attributes based on the specified order.
func sortAttributes(attributes map[string]*hclwrite.Attribute, order []string) []string {
	var sortedKeys []string
	orderIndex := make(map[string]int)

	for _, attr := range order {
		orderIndex[attr] = len(sortedKeys)
		sortedKeys = append(sortedKeys, attr)
	}

	for attr := range attributes {
		if _, exists := orderIndex[attr]; !exists {
			orderIndex[attr] = len(sortedKeys)
			sortedKeys = append(sortedKeys, attr)
		}
	}

	sort.SliceStable(sortedKeys, func(i, j int) bool {
		return orderIndex[sortedKeys[i]] < orderIndex[sortedKeys[j]]
	})

	return sortedKeys
}
