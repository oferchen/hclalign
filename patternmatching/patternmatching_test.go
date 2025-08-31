// patternmatching/patternmatching_test.go
package patternmatching_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/oferchen/hclalign/config"
	patternmatching "github.com/oferchen/hclalign/patternmatching"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatcherMatches(t *testing.T) {
	wd := t.TempDir()
	oldwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(wd))
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	require.NoError(t, os.WriteFile(filepath.Join(wd, "main.tf"), []byte(""), 0o644))
	require.NoError(t, os.Mkdir(filepath.Join(wd, "vendor"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(wd, "vendor", "v.tf"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(wd, "note.txt"), []byte(""), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(wd, "nested", "included"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(wd, "nested", "included", "in.tf"), []byte(""), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(wd, "nested", "excluded"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(wd, "nested", "excluded", "out.tf"), []byte(""), 0o644))

	m, err := patternmatching.NewMatcher([]string{"**/*.tf"}, []string{"**/vendor/**", "nested/excluded/**"}, wd)
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

func TestMatcherMatchesOutsideWorkingDir(t *testing.T) {
	wd := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(wd, "main.tf"), []byte(""), 0o644))
	require.NoError(t, os.Mkdir(filepath.Join(wd, "vendor"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(wd, "vendor", "v.tf"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(wd, "note.txt"), []byte(""), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(wd, "nested", "included"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(wd, "nested", "included", "in.tf"), []byte(""), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(wd, "nested", "excluded"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(wd, "nested", "excluded", "out.tf"), []byte(""), 0o644))

	m, err := patternmatching.NewMatcher([]string{"**/*.tf"}, []string{"**/vendor/**", "nested/excluded/**"}, wd)
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

func TestMatcherDefaultExclude(t *testing.T) {
	wd := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(wd, "main.tf"), []byte(""), 0o644))
	require.NoError(t, os.Mkdir(filepath.Join(wd, ".terraform"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(wd, ".terraform", "ignored.tf"), []byte(""), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(wd, "nested", ".terraform"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(wd, "nested", ".terraform", "included.tf"), []byte(""), 0o644))
	require.NoError(t, os.Mkdir(filepath.Join(wd, "vendor"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(wd, "vendor", "ignored.tf"), []byte(""), 0o644))

	m, err := patternmatching.NewMatcher(config.DefaultInclude, config.DefaultExclude, wd)
	require.NoError(t, err)

	assert.True(t, m.Matches(filepath.Join(wd, "main.tf")))
	assert.True(t, m.Matches(filepath.Join(wd, "nested", ".terraform")))
	assert.True(t, m.Matches(filepath.Join(wd, "nested", ".terraform", "included.tf")))
	paths := []string{
		filepath.Join(wd, ".terraform"),
		filepath.Join(wd, ".terraform", "ignored.tf"),
		filepath.Join(wd, "vendor"),
		filepath.Join(wd, "vendor", "ignored.tf"),
	}
	for _, p := range paths {
		assert.False(t, m.Matches(p))
	}
}

func TestValidatePatterns(t *testing.T) {
	err := patternmatching.ValidatePatterns([]string{"**/*.tf", "*.hcl"})
	assert.NoError(t, err)

	err = patternmatching.ValidatePatterns([]string{"["})
	assert.Error(t, err)

	err = patternmatching.ValidatePatterns([]string{""})
	assert.Error(t, err)
}

func TestNewMatcherInvalidPattern(t *testing.T) {
	_, err := patternmatching.NewMatcher([]string{"["}, nil, "")
	assert.Error(t, err)
}

func TestMatcherMatchesConcurrent(t *testing.T) {
	t.Parallel()

	wd := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(wd, "main.tf"), []byte(""), 0o644))

	m, err := patternmatching.NewMatcher([]string{"**/*.tf"}, nil, wd)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Matches(filepath.Join(wd, "main.tf"))
		}()
	}
	wg.Wait()
}

func TestMatcherRejectsOutsideRoot(t *testing.T) {
	root := t.TempDir()
	out := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(root, "in.tf"), []byte(""), 0o644))
	outsideFile := filepath.Join(out, "out.tf")
	require.NoError(t, os.WriteFile(outsideFile, []byte(""), 0o644))

	m, err := patternmatching.NewMatcher([]string{"**/*.tf"}, nil, root)
	require.NoError(t, err)

	assert.True(t, m.Matches(filepath.Join(root, "in.tf")))
	assert.False(t, m.Matches(outsideFile))

	upPath := filepath.Join(root, "..", filepath.Base(out), "out.tf")
	assert.False(t, m.Matches(upPath))
}
