//go:build windows

// filename: internal/fs/ewindows_windows.go
package fs

import (
	"errors"
	"syscall"
)

func isErrWindows(err error) bool {
	return errors.Is(err, syscall.EWINDOWS)
}
