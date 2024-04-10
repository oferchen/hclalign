package hclprocessing_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/oferchen/hclalign/fileprocessing"
	"github.com/oferchen/hclalign/hclprocessing"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func createTestHCLFile(attributesOrder []string) *hclwrite.File {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	block := rootBody.AppendNewBlock("resource", []string{"example", "test"})
	blockBody := block.Body()

	for _, attr := range attributesOrder {
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
			hclprocessing.ReorderAttributes(file, tt.desiredOrder)

			var resultOrder []string
			for attrName := range file.Body().Blocks()[0].Body().Attributes() {
				resultOrder = append(resultOrder, attrName)
			}
			require.ElementsMatch(t, tt.expectedOrder, resultOrder)
		})
	}
}

func TestProcessSingleFile_ValidHCL_PermissionsPreserved(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.hcl")
	initialContent := `variable "test" {
  attribute1 = "value1"
  attribute2 = "value2"
}`
	// Define custom permissions for the test file
	customPerms := fs.FileMode(0644)
	require.NoError(t, os.WriteFile(filePath, []byte(initialContent), customPerms))

	// Retrieve and store the original permissions of the file
	originalFileInfo, err := os.Stat(filePath)
	require.NoError(t, err)
	originalPerms := originalFileInfo.Mode()

	order := []string{"attribute2", "attribute1"}
	require.NoError(t, fileprocessing.ProcessSingleFile(filePath, order))

	// After processing, check that the file permissions have not changed
	processedFileInfo, err := os.Stat(filePath)
	require.NoError(t, err)
	require.Equal(t, originalPerms, processedFileInfo.Mode(), "File permissions should be preserved after processing")
}
