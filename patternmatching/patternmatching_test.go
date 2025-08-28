// patternmatching/patternmatching_test.go
package patternmatching

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatcherMatches(t *testing.T) {
	wd := t.TempDir()
	oldwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(wd))
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	require.NoError(t, os.WriteFile(filepath.Join(wd, "main.tf"), []byte(""), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(wd, "vendor"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(wd, "vendor", "v.tf"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(wd, "note.txt"), []byte(""), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(wd, "nested", "included"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(wd, "nested", "included", "in.tf"), []byte(""), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(wd, "nested", "excluded"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(wd, "nested", "excluded", "out.tf"), []byte(""), 0644))

	m, err := NewMatcher([]string{"**/*.tf"}, []string{"**/vendor/**", "nested/excluded/**"})
	require.NoError(t, err)

	assert.True(t, m.Matches(filepath.Join(wd, "main.tf")))
	assert.True(t, m.Matches(filepath.Join(wd, "nested", "included")))
	assert.True(t, m.Matches(filepath.Join(wd, "nested", "included", "in.tf")))
	assert.False(t, m.Matches(filepath.Join(wd, "note.txt")))
	assert.False(t, m.Matches(filepath.Join(wd, "vendor")))
	assert.False(t, m.Matches(filepath.Join(wd, "vendor", "v.tf")))
	assert.False(t, m.Matches(filepath.Join(wd, "nested", "excluded")))
	assert.False(t, m.Matches(filepath.Join(wd, "nested", "excluded", "out.tf")))
}

func TestValidatePatterns(t *testing.T) {
	err := ValidatePatterns([]string{"**/*.tf", "*.hcl"})
	assert.NoError(t, err)

	err = ValidatePatterns([]string{"["})
	assert.Error(t, err)
}
