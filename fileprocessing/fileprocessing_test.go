package fileprocessing_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/oferchen/hclalign/fileprocessing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDirectory(t *testing.T, files map[string]string) string {
	dir, err := os.MkdirTemp("", "hclprocessing")
	require.NoError(t, err, "creating temp directory should not fail")

	for path, content := range files {
		fullPath := filepath.Join(dir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755), "creating subdirectories should not fail")
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644), "writing test file should not fail")
	}

	return dir
}

func TestProcessFiles(t *testing.T) {
	testFiles := map[string]string{
		"valid1.hcl": `
resource "example" "test1" {
	attribute2 = "value2"
	attribute1 = "value1"
}`,
		"valid2.hcl": `
resource "example" "test2" {
	attribute3 = "value3"
	attribute2 = "value2"
}`,
		"subdir/ignored.txt": "This file should be ignored.",
	}

	dir := setupTestDirectory(t, testFiles)
	defer os.RemoveAll(dir)

	criteria := []string{"*.hcl"}
	order := []string{"attribute1", "attribute2", "attribute3"}

	err := fileprocessing.ProcessFiles(dir, criteria, order)
	require.NoError(t, err, "ProcessFiles should not return an error")

	// Validate the contents of the processed files
	for path := range testFiles {
		fullPath := filepath.Join(dir, path)
		content, err := os.ReadFile(fullPath)
		require.NoError(t, err, "reading processed file should not fail")

		if filepath.Ext(path) == ".hcl" {
			assert.Contains(t, string(content), "attribute1", "File content should contain attribute1")
			assert.Contains(t, string(content), "attribute2", "File content should contain attribute2")
		} else {
			assert.Equal(t, testFiles[path], string(content), "Ignored file content should not change")
		}
	}
}
