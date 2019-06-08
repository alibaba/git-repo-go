// +build windows

package cap

import (
	"errors"
)

func GetRlimitNoFile() (uint64, error) {
	return 0, errors.New("getrlimit not implement in Windows")
}
