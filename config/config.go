// filename: config/config.go
package config

import (
	"fmt"
	"runtime"
	"strings"

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
	Verbose            bool
	FollowSymlinks     bool
	ProvidersSchema    string
	UseTerraformSchema bool
	Types              []string
	PrefixOrder        bool
}

var (
	DefaultInclude = []string{"**/*.tf"}
	DefaultExclude = []string{"**/.terraform/**", "**/vendor/**", "**/.git/**", "**/node_modules/**"}
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
	if err := ValidateOrder(c.Order); err != nil {
		return fmt.Errorf("invalid order: %w", err)
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

func ValidateOrder(order []string) error {
	provided := make(map[string]struct{})
	for _, item := range order {
		if item == "" {
			return fmt.Errorf("attribute name cannot be empty")
		}
		if _, exists := provided[item]; exists {
			return fmt.Errorf("duplicate attribute '%s' found in order", item)
		}
		provided[item] = struct{}{}
	}
	return nil
}

func ParseOrder(order []string) ([]string, error) {
	attrs := make([]string, 0, len(order))
	attrSet := make(map[string]struct{}, len(order))
	for _, item := range order {
		if item == "" {
			return nil, fmt.Errorf("attribute name cannot be empty")
		}
		if strings.Contains(item, "=") {
			return nil, fmt.Errorf("unknown block ordering '%s'", item)
		}
		if _, exists := attrSet[item]; exists {
			return nil, fmt.Errorf("duplicate attribute '%s' found in order", item)
		}
		attrSet[item] = struct{}{}
		attrs = append(attrs, item)
	}
	return attrs, nil
}
