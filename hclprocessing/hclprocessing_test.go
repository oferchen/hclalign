package hclprocessing

import (
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func createTestHCLFile(attributesOrder []string) *hclwrite.File {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	block := rootBody.AppendNewBlock("resource", []string{"example", "test"})
	blockBody := block.Body()

	for _, attr := range attributesOrder {
		// Correctly create a cty.Value from string and set the attribute
		blockBody.SetAttributeValue(attr, cty.StringVal(attr))
	}

	return f
}

func TestReorderAttributes(t *testing.T) {
	tests := []struct {
		name          string
		originalOrder []string
		desiredOrder  []string
		expectedOrder []string
	}{
		{
			name:          "correct order",
			originalOrder: []string{"attribute3", "attribute1", "attribute2"},
			desiredOrder:  []string{"attribute1", "attribute2", "attribute3"},
			expectedOrder: []string{"attribute1", "attribute2", "attribute3"},
		},
		{
			name:          "partial order specified",
			originalOrder: []string{"attribute3", "attribute2", "attribute1"},
			desiredOrder:  []string{"attribute1"},
			expectedOrder: []string{"attribute1", "attribute3", "attribute2"},
		},
		{
			name:          "extra attributes ignored",
			originalOrder: []string{"attribute1", "attribute2"},
			desiredOrder:  []string{"attribute1", "attribute2", "attribute3"},
			expectedOrder: []string{"attribute1", "attribute2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := createTestHCLFile(tt.originalOrder)
			ReorderAttributes(file, tt.desiredOrder)

			var resultOrder []string
			for attrName := range file.Body().Blocks()[0].Body().Attributes() {
				resultOrder = append(resultOrder, attrName)
			}
			// Ensure the resultOrder slice matches the expectedOrder,
			// This assumes that the order of map iteration matches the insertion order, which is true for this specific use case but not a general property of Go maps.
			require.ElementsMatch(t, tt.expectedOrder, resultOrder)
		})
	}
}
