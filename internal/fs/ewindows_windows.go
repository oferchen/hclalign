//go:build windows

// internal/fs/ewindows_windows.go — SPDX-License-Identifier: Apache-2.0
package fs

import (
	"errors"
	"syscall"
)

func isErrWindows(err error) bool {
	return errors.Is(err, syscall.EWINDOWS)
}
