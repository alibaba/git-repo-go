package project

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/aliyun/git-repo-go/config"
)

const (
	groupDefaultConst    = "default"
	groupAllConst        = "all"
	groupNotDefaultConst = "notdefault"
)

var (
	emailUserPattern = regexp.MustCompile(`^.* <([^\s]+)@[^\s]+>$`)
)

// urlJoin appends fetch path (in remote element) and project name to manifest url.
func urlJoin(u string, paths ...string) (string, error) {
	var err error

	// remove last part of url
	if len(u) > 0 && u[len(u)-1] == '/' {
		u = u[0 : len(u)-1]
	}
	i := strings.LastIndex(u, "/")
	if i > 0 {
		u = u[0:i]
	}

	for _, p := range paths {
		u, err = joinTwoURL(u, p)
		if err != nil {
			return "", err
		}
	}
	return u, nil
}

func joinTwoURL(u, p string) (string, error) {
	var (
		prefix    string
		remain    string
		keepSlash bool
	)

	if filepath.IsAbs(p) || strings.Contains(p, ":") {
		return p, nil
	}

	if strings.Contains(u, "://") {
		slices := strings.SplitN(u, "://", 2)
		prefix = slices[0] + "://"
		remains := strings.SplitN(slices[1], "/", 2)
		prefix += remains[0] + "/"
		if len(remains) == 1 {
			remain = ""
		} else {
			remain = remains[1]
		}
	} else if strings.Contains(u, ":") {
		slices := strings.SplitN(u, ":", 2)
		prefix = slices[0] + ":"
		remain = slices[1]
	} else if filepath.IsAbs(u) {
		prefix = "/"
		remain = u[1:]
	} else {
		return "", fmt.Errorf("invalid git url: %s", u)
	}

	if len(remain) == 0 {
		remain = "/"
	} else if remain[0] == '/' {
		keepSlash = true
	} else {
		remain = "/" + remain
	}

	remain = filepath.Join(remain, p)
	if !keepSlash && len(remain) > 0 && remain[0] == '/' {
		remain = remain[1:]
	}
	return prefix + remain, nil
}

// MatchGroups checks if project has matched groups.
func MatchGroups(match, groups string) bool {
	matchGroups := []string{}
	for _, g := range strings.Split(match, ",") {
		matchGroups = append(matchGroups, strings.TrimSpace(g))
	}
	if len(matchGroups) == 0 {
		matchGroups = append(matchGroups, groupDefaultConst)
	}

	projectGroups := []string{groupAllConst}
	hasNotDefault := false
	for _, g := range strings.Split(groups, ",") {
		g = strings.TrimSpace(g)
		projectGroups = append(projectGroups, g)
		if g == groupNotDefaultConst {
			hasNotDefault = true
		}
	}
	if !hasNotDefault {
		projectGroups = append(projectGroups, groupDefaultConst)
	}

	matched := false
	for _, g := range matchGroups {
		inverse := false
		if strings.HasPrefix(g, "-") {
			inverse = true
			g = g[1:]
		}
		for _, pg := range projectGroups {
			if pg == g {
				matched = !inverse
				break
			}
		}
	}

	return matched

}

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
