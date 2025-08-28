// hclprocessing.go
// Manages parsing and reordering of attributes within HCL files.

package hclprocessing

import (
	"sort"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

// ReorderAttributes reorders the attributes of the given HCL file according to the specified order.
func ReorderAttributes(file *hclwrite.File, order []string) {
	body := file.Body()
	blocks := body.Blocks()
	for _, b := range blocks {
		body.RemoveBlock(b)
	}

	reorderBodyAttributes(body, order)

	for _, b := range blocks {
		body.AppendBlock(b)
		reorderBlockAttributes(b, order)
	}
}

// reorderBlockAttributes reorders attributes within a single block based on the specified order.
func reorderBlockAttributes(block *hclwrite.Block, order []string) {
	body := block.Body()
	nested := body.Blocks()
	for _, b := range nested {
		body.RemoveBlock(b)
	}

	reorderBodyAttributes(body, order)

	for _, b := range nested {
		body.AppendBlock(b)
		reorderBlockAttributes(b, order)
	}
}

// reorderBodyAttributes reorders attributes in the given body based on the specified order.
func reorderBodyAttributes(body *hclwrite.Body, order []string) {
	originalAttributes := body.Attributes()
	sortedAttributes := sortAttributes(originalAttributes, order)

	// Remove all attributes to re-add them in the correct order.
	for name := range originalAttributes {
		body.RemoveAttribute(name)
	}

	// Re-add attributes in the specified order.
	for _, name := range sortedAttributes {
		if attr, exists := originalAttributes[name]; exists {
			body.SetAttributeRaw(name, attr.Expr().BuildTokens(nil))
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
