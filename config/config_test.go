// filename: config/config_test.go
package config

import (
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestValidateOrder_UnknownAttributeAllowed(t *testing.T) {
	order := append([]string{}, CanonicalOrder...)
	order = append(order, "unknown")
	if err := ValidateOrder(order); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateOrder_EmptyAttribute(t *testing.T) {
	err := ValidateOrder([]string{""})
	if err == nil || !strings.Contains(err.Error(), "attribute name cannot be empty") {
		t.Fatalf("expected error for empty attribute name, got %v", err)
	}
}

func TestCanonicalOrderMatchesBuiltInAttributes(t *testing.T) {
	expected := []string{"description", "type", "default", "sensitive", "nullable"}
	if !reflect.DeepEqual(CanonicalOrder, expected) {
		t.Fatalf("expected CanonicalOrder to be %v, got %v", expected, CanonicalOrder)
	}
}

func TestValidateOrder_EmptyAttributeName(t *testing.T) {
	err := ValidateOrder([]string{"description", ""})
	if err == nil || err.Error() != "attribute name cannot be empty" {
		t.Fatalf("expected error for empty attribute name, got %v", err)
	}
}

func TestDefaultExcludeMatchesExpected(t *testing.T) {
	expected := []string{".terraform/**", "**/.terraform/**", "vendor/**"}
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
