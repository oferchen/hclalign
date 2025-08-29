// internal/fs/ewindows_windows.go
//go:build windows

package fs

import (
	"errors"
	"syscall"
)

func isErrWindows(err error) bool {
	return errors.Is(err, syscall.EWINDOWS)
}
