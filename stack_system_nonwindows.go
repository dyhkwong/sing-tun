//go:build !windows

package tun

import (
	"errors"

	"golang.org/x/sys/unix"
)

func retryableListenError(err error) bool {
	return errors.Is(err, unix.EADDRNOTAVAIL)
}
