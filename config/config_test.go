// config/config_test.go
package config

import (
	"reflect"
	"runtime"
	"testing"
)

func TestCanonicalOrderMatchesBuiltInAttributes(t *testing.T) {
	expected := []string{"description", "type", "default", "sensitive", "nullable"}
	if !reflect.DeepEqual(CanonicalOrder, expected) {
		t.Fatalf("expected CanonicalOrder to be %v, got %v", expected, CanonicalOrder)
	}
}

func TestDefaultIncludeMatchesExpected(t *testing.T) {
	expected := []string{"**/*.tf"}
	if !reflect.DeepEqual(DefaultInclude, expected) {
		t.Fatalf("expected DefaultInclude to be %v, got %v", expected, DefaultInclude)
	}
}

func TestDefaultExcludeMatchesExpected(t *testing.T) {
	expected := []string{".terraform/**", "vendor/**"}
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
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateTypes(t *testing.T) {
	c := Config{Concurrency: 1, Include: DefaultInclude, Exclude: DefaultExclude, Types: []string{"", "var"}}
	if err := c.Validate(); err == nil {
		t.Fatalf("expected error for empty type name")
	}
	c = Config{Concurrency: 1, Include: DefaultInclude, Exclude: DefaultExclude, Types: []string{"a", "a"}}
	if err := c.Validate(); err == nil {
		t.Fatalf("expected error for duplicate type")
	}
	c = Config{Concurrency: 1, Include: DefaultInclude, Exclude: DefaultExclude, Types: []string{}}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Types) != 1 || c.Types[0] != "variable" {
		t.Fatalf("unexpected default type: %v", c.Types)
	}
}
