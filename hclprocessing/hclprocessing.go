// hclprocessing.go
// Manages parsing and reordering of attributes within HCL files.

package hclprocessing

import (
	"sort"

	"github.com/hashicorp/hcl/v2/hclwrite"
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
// This version improves on handling attributes not explicitly mentioned in the order.
func sortAttributes(attributes map[string]*hclwrite.Attribute, order []string) []string {
	orderMap := make(map[string]int)
	for i, attrName := range order {
		orderMap[attrName] = i
	}

	var sortedKeys []string
	for attr := range attributes {
		sortedKeys = append(sortedKeys, attr)
	}

	sort.Slice(sortedKeys, func(i, j int) bool {
		indexI, foundI := orderMap[sortedKeys[i]]
		indexJ, foundJ := orderMap[sortedKeys[j]]

		if foundI && foundJ {
			return indexI < indexJ // Both are in order, sort by order index
		}
		if foundI || foundJ {
			return foundI // If only one is found, it comes first
		}
		return sortedKeys[i] < sortedKeys[j] // Neither in order, sort alphabetically
	})

	return sortedKeys
}
