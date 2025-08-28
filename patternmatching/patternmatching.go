package patternmatching

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

// Matcher evaluates include and exclude glob patterns against file paths.
type Matcher struct {
	include []string
	exclude []string
	root    string
}

// NewMatcher creates a Matcher using include and exclude patterns relative to the
// current working directory.
func NewMatcher(include, exclude []string) (*Matcher, error) {
	if err := validatePatterns(include); err != nil {
		return nil, fmt.Errorf("invalid include: %w", err)
	}
	if err := validatePatterns(exclude); err != nil {
		return nil, fmt.Errorf("invalid exclude: %w", err)
	}
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return &Matcher{include: include, exclude: exclude, root: root}, nil
}

// validatePatterns ensures each glob pattern is syntactically valid.
func validatePatterns(patterns []string) error {
	for _, p := range patterns {
		if p == "" {
			return fmt.Errorf("invalid pattern: pattern is empty")
		}
		if _, err := doublestar.PathMatch(p, ""); err != nil {
			return fmt.Errorf("invalid pattern '%s': %w", p, err)
		}
	}
	return nil
}

// Matches reports whether the given path should be processed. It returns false
// for files that do not match the include patterns or that match the exclude
// patterns. Directories are always matched unless they are excluded.
func (m *Matcher) Matches(path string) bool {
	rel, err := filepath.Rel(m.root, path)
	if err != nil {
		rel = path
	}
	// Normalize path separators for consistent glob matching across platforms.
	rel = filepath.ToSlash(rel)
	// Check excludes first.
	for _, ex := range m.exclude {
		if ok, _ := doublestar.PathMatch(ex, rel); ok {
			return false
		}
	}
	info, err := os.Stat(path)
	isDir := err == nil && info.IsDir()
	if isDir {
		return true
	}
	for _, in := range m.include {
		if ok, _ := doublestar.PathMatch(in, rel); ok {
			return true
		}
	}
	return false
}

// ValidatePatterns is exported for configuration validation.
func ValidatePatterns(patterns []string) error { return validatePatterns(patterns) }
