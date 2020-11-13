package common

import (
	"strings"
	"unicode"

	"github.com/alibaba/git-repo-go/config"
)

// IsSha indecates revision is a commit id (hash).
func IsSha(revision string) bool {
	if config.CommitIDPattern.MatchString(revision) {
		return true
	}
	return false
}

// IsTag indecates revision is a tag.
func IsTag(revision string) bool {
	if strings.HasPrefix(revision, config.RefsTags) {
		return true
	}
	return false
}

// IsHead indecates revision is a branch.
func IsHead(revision string) bool {
	if strings.HasPrefix(revision, config.RefsHeads) {
		return true
	}
	return false
}

// IsRef indecates revision is a ref.
func IsRef(revision string) bool {
	if strings.HasPrefix(revision, "refs/") {
		return true
	}
	return false
}

// IsImmutable indecates revision is a tag or sha or change.
func IsImmutable(revision string) bool {
	if IsSha(revision) || IsTag(revision) {
		return true
	}
	if IsHead(revision) {
		return false
	}
	if strings.HasPrefix(revision, "refs/") {
		return true
	}
	// is a head
	return false
}

// IsASCII indicates string contains only ASCII.
func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
