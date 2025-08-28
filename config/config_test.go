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

func TestValidateOrder_StrictDefaultOrder(t *testing.T) {
	if err := ValidateOrder(DefaultOrder, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefaultOrderMatchesBuiltInAttributes(t *testing.T) {
	expected := []string{"description", "type", "default", "sensitive", "nullable", "validation"}
	if !reflect.DeepEqual(DefaultOrder, expected) {
		t.Fatalf("expected DefaultOrder to be %v, got %v", expected, DefaultOrder)
	}
}

func TestDefaultExcludeMatchesExpected(t *testing.T) {
	expected := []string{"**/.terraform/**", "**/vendor/**", "**/.git/**", "**/node_modules/**"}
	if !reflect.DeepEqual(DefaultExclude, expected) {
		t.Fatalf("expected DefaultExclude to be %v, got %v", expected, DefaultExclude)
	}
}
