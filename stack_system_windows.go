package tun

import (
	"errors"

	"golang.org/x/sys/windows"
)

func retryableListenError(err error) bool {
	return errors.Is(err, windows.WSAEADDRNOTAVAIL)
}
