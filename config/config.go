// /config/config.go
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
	StrictOrder        bool
	Concurrency        int
	Verbose            bool
	FollowSymlinks     bool
	FmtOnly            bool
	NoFmt              bool
	FmtStrategy        string
	ProvidersSchema    string
	UseTerraformSchema bool
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
	if c.FmtOnly && c.NoFmt {
		return fmt.Errorf("fmt-only and no-fmt cannot be used together")
	}
	if err := patternmatching.ValidatePatterns(c.Include); err != nil {
		return fmt.Errorf("invalid include: %w", err)
	}
	if err := patternmatching.ValidatePatterns(c.Exclude); err != nil {
		return fmt.Errorf("invalid exclude: %w", err)
	}
	if err := ValidateOrder(c.Order, c.StrictOrder); err != nil {
		return fmt.Errorf("invalid order: %w", err)
	}
	return nil
}

func ValidateOrder(order []string, strict bool) error {
	providedCanonical := make(map[string]struct{})
	providedFlags := make(map[string]struct{})
	canonicalSet := make(map[string]struct{}, len(CanonicalOrder))
	for _, item := range CanonicalOrder {
		canonicalSet[item] = struct{}{}
	}

	for _, item := range order {
		if item == "" {
			return fmt.Errorf("attribute name cannot be empty")
		}
		if strings.Contains(item, "=") {
			block, val, ok := strings.Cut(item, "=")
			if !ok || block == "" || val == "" {
				return fmt.Errorf("invalid block ordering '%s'", item)
			}
			if block != "locals" || val != "alphabetical" {
				return fmt.Errorf("unknown block ordering '%s'", item)
			}
			if _, exists := providedFlags[block]; exists {
				return fmt.Errorf("duplicate attribute '%s' found in order", item)
			}
			providedFlags[block] = struct{}{}
			continue
		}
		if _, exists := providedCanonical[item]; exists {
			return fmt.Errorf("duplicate attribute '%s' found in order", item)
		}
		providedCanonical[item] = struct{}{}
	}

	if strict {
		for item := range providedCanonical {
			if _, ok := canonicalSet[item]; !ok {
				return fmt.Errorf("unknown attribute '%s' in order", item)
			}
		}
		for _, item := range CanonicalOrder {
			if _, exists := providedCanonical[item]; !exists {
				return fmt.Errorf("missing expected attribute '%s' in provided order", item)
			}
		}
	}
	return nil
}
