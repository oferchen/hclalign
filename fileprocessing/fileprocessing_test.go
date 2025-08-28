package fileprocessing_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"io/fs"
	"strings"

	"github.com/oferchen/hclalign/fileprocessing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to setup a test environment with HCL files.
func setupTestDir(t *testing.T, files map[string]string) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "hcltest")
	require.NoError(t, err)

	for name, content := range files {
		fullPath := filepath.Join(tmpDir, name)
		err := os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}

	return tmpDir
}

func TestProcessFiles(t *testing.T) {
	t.Run("process matching files", func(t *testing.T) {
		files := map[string]string{
			"match1.hcl": `attribute1 = "value1"
attribute2 = "value2"`,
			"match2.hcl": `attribute3 = "value3"
attribute4 = "value4"`,
			"nomatch.txt": "This should not be modified.",
		}
		tmpDir := setupTestDir(t, files)
		defer os.RemoveAll(tmpDir)

		criteria := []string{`.*\.hcl$`}
		order := []string{"attribute2", "attribute1", "attribute4", "attribute3"}

		err := fileprocessing.ProcessFiles(context.Background(), tmpDir, criteria, order)
		require.NoError(t, err)

		// Verify contents of the files after processing
		err = filepath.WalkDir(tmpDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				content, err := os.ReadFile(path)
				require.NoError(t, err)
				if strings.HasSuffix(path, "match1.hcl") {
					assert.Contains(t, string(content), `attribute2 = "value2"`)
				} else if strings.HasSuffix(path, "match2.hcl") {
					assert.Contains(t, string(content), `attribute4 = "value4"`)
				} else if strings.HasSuffix(path, "nomatch.txt") {
					assert.Equal(t, "This should not be modified.", string(content))
				}
			}
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("empty directory", func(t *testing.T) {
		tmpDir := setupTestDir(t, map[string]string{})
		defer os.RemoveAll(tmpDir)

		err := fileprocessing.ProcessFiles(context.Background(), tmpDir, []string{`.*\.hcl$`}, []string{"attribute1", "attribute2"})
		assert.NoError(t, err)
	})

	t.Run("invalid criteria", func(t *testing.T) {
		tmpDir := setupTestDir(t, map[string]string{"test.hcl": `attribute = "value"`})
		defer os.RemoveAll(tmpDir)

		err := fileprocessing.ProcessFiles(context.Background(), tmpDir, []string{"[\\"}, []string{"attribute"})
		assert.Error(t, err)
	})
}

func TestProcessSingleFile_ValidHCL(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.hcl")
	initialContent := `variable "test" {
  attribute1 = "value1"
  attribute2 = "value2"
}`
	require.NoError(t, os.WriteFile(filePath, []byte(initialContent), 0644))

	order := []string{"attribute2", "attribute1"}
	require.NoError(t, fileprocessing.ProcessSingleFile(filePath, order))

	resultContent, err := os.ReadFile(filePath)
	require.NoError(t, err)

	// Assuming the ReorderAttributes function places unmatched attributes alphabetically at the end.
	expectedContent := `variable "test" {
  attribute2 = "value2"
  attribute1 = "value1"
}`
	assert.Equal(t, expectedContent, string(resultContent), "Attribute order is incorrect.")
}

// Ensure non-HCL files are skipped without error
func TestProcessSingleFile_NonHCLContent(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	nonHCLContent := "This is not HCL content."

	require.NoError(t, os.WriteFile(filePath, []byte(nonHCLContent), 0644))

	// Process the file; expect an error because the content is not valid HCL
	err := fileprocessing.ProcessSingleFile(filePath, []string{})
	require.Error(t, err, "Processing non-HCL content should result in an error")
	require.Contains(t, err.Error(), "parsing error", "The error message should indicate a parsing error")

	// Verify the file content remains unchanged
	resultContent, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, nonHCLContent, string(resultContent))
}
