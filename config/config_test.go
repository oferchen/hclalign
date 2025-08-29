// /config/config_test.go
package config

import (
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestValidateOrder_StrictUnknownAttribute(t *testing.T) {
	err := ValidateOrder([]string{"description", "unknown"}, true)
	if err == nil {
		t.Fatalf("expected error for unknown attribute in strict mode")
	}
}

func TestValidateOrder_NonStrictUnknownAttribute(t *testing.T) {
	if err := ValidateOrder([]string{"description", "unknown"}, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateOrder_StrictCanonicalOrder(t *testing.T) {
	if err := ValidateOrder(CanonicalOrder, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateOrder_StrictMissingAttribute(t *testing.T) {
	subset := CanonicalOrder[:len(CanonicalOrder)-1]
	err := ValidateOrder(subset, true)
	if err == nil || !strings.Contains(err.Error(), "missing expected attribute") {
		t.Fatalf("expected missing expected attribute error, got %v", err)
	}
}

func TestValidateOrder_EmptyAttribute(t *testing.T) {
	err := ValidateOrder([]string{""}, false)
	if err == nil || !strings.Contains(err.Error(), "attribute name cannot be empty") {
		t.Fatalf("expected error for empty attribute name, got %v", err)
	}
}

func TestValidateOrder_BlockOrderingFlag(t *testing.T) {
	if err := ValidateOrder([]string{"locals=alphabetical"}, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateOrder_DuplicateBlockOrderingFlag(t *testing.T) {
	err := ValidateOrder([]string{"locals=alphabetical", "locals=alphabetical"}, false)
	if err == nil {
		t.Fatalf("expected error for duplicate block ordering flag")
	}
}

func TestCanonicalOrderMatchesBuiltInAttributes(t *testing.T) {
	expected := []string{"description", "type", "default", "sensitive", "nullable"}
	if !reflect.DeepEqual(CanonicalOrder, expected) {
		t.Fatalf("expected CanonicalOrder to be %v, got %v", expected, CanonicalOrder)
	}
}

func TestValidateOrder_StrictValidationBlock(t *testing.T) {
	if err := ValidateOrder([]string{"description", "validation"}, true); err == nil {
		t.Fatalf("expected error for validation block in strict mode")
	}
}

func TestValidateOrder_EmptyAttributeName(t *testing.T) {
	err := ValidateOrder([]string{"description", ""}, false)
	if err == nil || err.Error() != "attribute name cannot be empty" {
		t.Fatalf("expected error for empty attribute name, got %v", err)
	}
}

func TestDefaultExcludeMatchesExpected(t *testing.T) {
	expected := []string{"**/.terraform/**", "**/vendor/**", "**/.git/**", "**/node_modules/**"}
	if !reflect.DeepEqual(DefaultExclude, expected) {
		t.Fatalf("expected DefaultExclude to be %v, got %v", expected, DefaultExclude)
	}
}

func TestValidate_ConcurrencyLessThanOne(t *testing.T) {
	c := Config{Concurrency: 0}
	if err := c.Validate(); err == nil {
		t.Fatalf("expected error for concurrency <1")
	}
}

func TestValidate_ConcurrencyGreaterThanGOMAXPROCS(t *testing.T) {
	c := Config{Concurrency: runtime.GOMAXPROCS(0) + 1}
	if err := c.Validate(); err == nil {
		t.Fatalf("expected error for concurrency > GOMAXPROCS")
	}
}

func TestValidate_InvalidIncludePattern(t *testing.T) {
	c := Config{Concurrency: 1, Include: []string{"["}}
	if err := c.Validate(); err == nil {
		t.Fatalf("expected error for invalid include pattern")
	}
}

func TestValidate_InvalidExcludePattern(t *testing.T) {
	c := Config{Concurrency: 1, Exclude: []string{"["}}
	if err := c.Validate(); err == nil {
		t.Fatalf("expected error for invalid exclude pattern")
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	c := Config{
		Concurrency: 1,
		Include:     DefaultInclude,
		Exclude:     DefaultExclude,
		Order:       CanonicalOrder,
		StrictOrder: true,
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidate_DuplicateOrderAttribute(t *testing.T) {
	c := Config{
		Concurrency: 1,
		Include:     DefaultInclude,
		Exclude:     DefaultExclude,
		Order:       []string{"description", "description"},
	}
	if err := c.Validate(); err == nil {
		t.Fatalf("expected error for duplicate order attribute")
	}
}

func TestValidate_EmptyOrderAttribute(t *testing.T) {
	c := Config{
		Concurrency: 1,
		Include:     DefaultInclude,
		Exclude:     DefaultExclude,
		Order:       []string{""},
	}
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "attribute name cannot be empty") {
		t.Fatalf("expected error for empty order attribute, got %v", err)
	}
}
