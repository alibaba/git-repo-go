// +build !windows

package cap

import (
	"syscall"
)

// GetRlimitNoFile gets syscall.RLIMIT_NOFILE
func GetRlimitNoFile() (uint64, error) {
	rlimit := syscall.Rlimit{}
	syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit)
	return rlimit.Cur, nil
}
