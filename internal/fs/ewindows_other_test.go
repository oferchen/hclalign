// internal/fs/ewindows_other_test.go
package fs

import (
	"errors"
	"testing"
)

func TestIsErrWindows(t *testing.T) {
	if isErrWindows(errors.New("boom")) {
		t.Fatalf("expected false")
	}
}
