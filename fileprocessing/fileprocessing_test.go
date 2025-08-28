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
			"match1.hcl": `variable "v1" {
  default     = "v1"
  description = "d1"
  extra       = true
}`,
			"match2.hcl": `variable "v2" {
  default     = "v2"
  description = "d2"
}`,
			"nomatch.txt": "This should not be modified.",
		}
		tmpDir := setupTestDir(t, files)
		defer os.RemoveAll(tmpDir)

		criteria := []string{"*.hcl"}

		err := fileprocessing.ProcessFiles(context.Background(), tmpDir, criteria, nil)
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
					assert.Equal(t, `variable "v1" {
  description = "d1"
  default     = "v1"
  extra       = true
}`, string(content))
				} else if strings.HasSuffix(path, "match2.hcl") {
					assert.Equal(t, `variable "v2" {
  description = "d2"
  default     = "v2"
}`, string(content))
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

		err := fileprocessing.ProcessFiles(context.Background(), tmpDir, []string{"*.hcl"}, nil)
		assert.NoError(t, err)
	})

	// Validation of criteria is handled outside ProcessFiles; invalid patterns are tested elsewhere.
}

func TestProcessSingleFile_ValidHCL(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.hcl")
	initialContent := `variable "test" {
  default     = "value1"
  description = "value2"
}`
	require.NoError(t, os.WriteFile(filePath, []byte(initialContent), 0644))

	require.NoError(t, fileprocessing.ProcessSingleFile(context.Background(), filePath, nil))

	resultContent, err := os.ReadFile(filePath)
	require.NoError(t, err)

	expectedContent := `variable "test" {
  description = "value2"
  default     = "value1"
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
	err := fileprocessing.ProcessSingleFile(context.Background(), filePath, nil)
	require.Error(t, err, "Processing non-HCL content should result in an error")
	require.Contains(t, err.Error(), "parsing error", "The error message should indicate a parsing error")

	// Verify the file content remains unchanged
	resultContent, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, nonHCLContent, string(resultContent))
}

func TestProcessFiles_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join("testdata", "idempotent_input.hcl")
	goldenPath := filepath.Join("testdata", "idempotent_golden.hcl")

	inputBytes, err := os.ReadFile(inputPath)
	require.NoError(t, err)
	testFile := filepath.Join(tmpDir, "test.hcl")
	require.NoError(t, os.WriteFile(testFile, inputBytes, 0644))

	criteria := []string{"*.hcl"}

	// First run
	require.NoError(t, fileprocessing.ProcessFiles(context.Background(), tmpDir, criteria, nil))
	expected, err := os.ReadFile(goldenPath)
	require.NoError(t, err)
	result1, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Equal(t, string(expected), string(result1))

	// Second run should be idempotent
	require.NoError(t, fileprocessing.ProcessFiles(context.Background(), tmpDir, criteria, nil))
	result2, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Equal(t, string(expected), string(result2))
}
