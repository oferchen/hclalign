package config

import "testing"

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
