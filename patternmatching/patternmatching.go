// Package patternmatching provides utilities for matching file paths against
// include and exclude glob patterns. Matcher values are safe for concurrent use
// by multiple goroutines.
package patternmatching

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
)

type Matcher struct {
	include  []string
	exclude  []string
	rootOnce sync.Once
	root     string
}

func NewMatcher(include, exclude []string) (*Matcher, error) {
	if err := validatePatterns(include); err != nil {
		return nil, fmt.Errorf("invalid include: %w", err)
	}
	if err := validatePatterns(exclude); err != nil {
		return nil, fmt.Errorf("invalid exclude: %w", err)
	}
	return &Matcher{include: include, exclude: exclude}, nil
}

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

func (m *Matcher) Matches(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	m.rootOnce.Do(func() {
		info, err := os.Stat(absPath)
		if err == nil && info.IsDir() {
			m.root = absPath
		} else {
			m.root = filepath.Dir(absPath)
		}
	})
	relPath, err := filepath.Rel(m.root, absPath)
	if err != nil {
		relPath = absPath
	}

	relPath = filepath.ToSlash(relPath)

	for _, ex := range m.exclude {
		if ok, _ := doublestar.PathMatch(ex, relPath); ok {
			return false
		}
	}
	info, err := os.Stat(path)
	isDir := err == nil && info.IsDir()
	if isDir {
		return true
	}
	for _, in := range m.include {
		if ok, _ := doublestar.PathMatch(in, relPath); ok {
			return true
		}
	}
	return false
}

func ValidatePatterns(patterns []string) error { return validatePatterns(patterns) }
