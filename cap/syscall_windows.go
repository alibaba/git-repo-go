// +build windows

package cap

import (
	"errors"
)

// GetRlimitNoFile() implements nothing, but returns error on Windows.
func GetRlimitNoFile() (uint64, error) {
	return 0, errors.New("getrlimit not implement in Windows")
}
