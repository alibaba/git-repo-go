package versions

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"

	"github.com/jiangxin/multi-log"
)

// Macros for version package
const (
	MinGitVersion = "1.7.2"
)

var (
	// Version of git-repo
	Version = "undefined"
)

// GetVersion show git-repo version
func GetVersion() string {
	return Version
}

// GetGitVersion gets current installed git version
func GetGitVersion() string {
	var out bytes.Buffer

	cmd := exec.Command("git", "version")
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Errorf("fail to run git version")
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(out.String(), "git version "))
}

// CompareVersion compares two versions
func CompareVersion(_left, _right string) int {
	left := strings.Split(_left, ".")
	right := strings.Split(_right, ".")
	pos := len(left)
	if pos > len(right) {
		pos = len(right)
	}

	for i := 0; i < pos; i++ {
		l, lErr := strconv.Atoi(left[i])
		r, rErr := strconv.Atoi(right[i])

		if lErr != nil && rErr != nil {
			if left[i] > right[i] {
				return 1
			} else if left[i] < right[i] {
				return -1
			} else {
				continue
			}
		} else if lErr != nil {
			return -1
		} else if rErr != nil {
			return 1
		}

		if l > r {
			return 1
		} else if l < r {
			return -1
		}
	}

	if len(left) > len(right) {
		if _, err := strconv.Atoi(left[pos]); err == nil {
			return 1
		}
		return -1
	} else if len(left) < len(right) {
		if _, err := strconv.Atoi(right[pos]); err == nil {
			return -1
		}
		return 1
	}

	return 0
}

// ValidateGitVersion is used to check installed git version
func ValidateGitVersion() bool {
	return CompareVersion(GetGitVersion(), MinGitVersion) >= 0
}
