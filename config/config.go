// config/config.go
package config

import (
	"fmt"
	"runtime"

	"github.com/oferchen/hclalign/patternmatching"
)

type Mode int

const (
	ModeWrite Mode = iota

	ModeCheck

	ModeDiff
)

type Config struct {
	Target             string
	Mode               Mode
	Stdin              bool
	Stdout             bool
	Include            []string
	Exclude            []string
	Order              []string
	Concurrency        int
	ProvidersSchema    string
	UseTerraformSchema bool
	Types              []string
	FollowSymlinks     bool
}

var (
	DefaultInclude = []string{"**/*.tf"}
	DefaultExclude = []string{".terraform/**", "vendor/**"}
	CanonicalOrder = []string{"description", "type", "default", "sensitive", "nullable"}
)

const (
	ErrMissingTarget = "missing target file or directory. Please provide a valid target as an argument"
)

func (c *Config) Validate() error {
	if c.Concurrency < 1 {
		return fmt.Errorf("concurrency must be at least 1")
	}
	if c.Concurrency > runtime.GOMAXPROCS(0) {
		return fmt.Errorf("concurrency cannot exceed GOMAXPROCS (%d)", runtime.GOMAXPROCS(0))
	}
	if err := patternmatching.ValidatePatterns(c.Include); err != nil {
		return fmt.Errorf("invalid include: %w", err)
	}
	if err := patternmatching.ValidatePatterns(c.Exclude); err != nil {
		return fmt.Errorf("invalid exclude: %w", err)
	}
	if c.Types != nil {
		if len(c.Types) == 0 {
			c.Types = []string{"variable"}
		}
		seen := make(map[string]struct{}, len(c.Types))
		for _, t := range c.Types {
			if t == "" {
				return fmt.Errorf("type name cannot be empty")
			}
			if _, ok := seen[t]; ok {
				return fmt.Errorf("duplicate type '%s'", t)
			}
			seen[t] = struct{}{}
		}
	}
	return nil
}
