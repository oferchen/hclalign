package fileprocessing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"io/fs"
	"strings"
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

		err := ProcessFiles(tmpDir, criteria, order)
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

		err := ProcessFiles(tmpDir, []string{`.*\.hcl$`}, []string{"attribute1", "attribute2"})
		assert.NoError(t, err)
	})

	t.Run("invalid criteria", func(t *testing.T) {
		tmpDir := setupTestDir(t, map[string]string{"test.hcl": `attribute = "value"`})
		defer os.RemoveAll(tmpDir)

		err := ProcessFiles(tmpDir, []string{"[\\"}, []string{"attribute"})
		assert.Error(t, err)
	})
}

func TestProcessSingleFile(t *testing.T) {
	t.Run("process valid HCL file", func(t *testing.T) {
		// Setup temporary file with initial HCL content
		tmpDir, err := os.MkdirTemp("", "hcltest")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		filePath := filepath.Join(tmpDir, "test.hcl")
		initialContent := `attribute1 = "value1"
attribute2 = "value2"`
		require.NoError(t, os.WriteFile(filePath, []byte(initialContent), 0644))

		order := []string{"attribute2", "attribute1"}

		// Process the file
		err = processSingleFile(filePath, order)
		require.NoError(t, err)

		// Read the file back and check the order
		resultContent, err := os.ReadFile(filePath)
		require.NoError(t, err)

		expectedContent := `attribute2 = "value2"
attribute1 = "value1"`
		assert.Equal(t, expectedContent, string(resultContent))
	})

	t.Run("file does not exist", func(t *testing.T) {
		err := processSingleFile("/path/to/nonexistent/file", []string{})
		assert.Error(t, err)
	})

	t.Run("invalid HCL content", func(t *testing.T) {
		// Setup temporary file with invalid HCL content
		tmpDir, err := os.MkdirTemp("", "hcltest")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		filePath := filepath.Join(tmpDir, "invalid.hcl")
		invalidContent := `attribute1 = "value1"
attribute2 "value2"` // Missing equals sign
		require.NoError(t, os.WriteFile(filePath, []byte(invalidContent), 0644))

		order := []string{"attribute2", "attribute1"}

		// Process the file
		err = processSingleFile(filePath, order)
		assert.Error(t, err)
	})

	t.Run("non-HCL file", func(t *testing.T) {
		// Setup temporary file with non-HCL content that should remain unchanged
		tmpDir, err := os.MkdirTemp("", "hcltest")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		filePath := filepath.Join(tmpDir, "test.txt")
		nonHCLContent := `This is not HCL content.`
		require.NoError(t, os.WriteFile(filePath, []byte(nonHCLContent), 0644))

		order := []string{} // No specific order

		// Attempt to process the file
		err = processSingleFile(filePath, order)
		require.NoError(t, err) // Expect no error since the function only processes HCL files

		// Read the file back and ensure content is unchanged
		resultContent, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, nonHCLContent, string(resultContent))
	})
}
