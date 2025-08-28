// config/config_test.go
package config

import (
	"reflect"
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

func TestDefaultExcludeMatchesExpected(t *testing.T) {
	expected := []string{"**/.terraform/**", "**/vendor/**", "**/.git/**", "**/node_modules/**"}
	if !reflect.DeepEqual(DefaultExclude, expected) {
		t.Fatalf("expected DefaultExclude to be %v, got %v", expected, DefaultExclude)
	}
}
