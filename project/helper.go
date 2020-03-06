package project

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/alibaba/git-repo-go/config"
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

func joinTwoURL(l, r string) (string, error) {
	lURL := config.ParseGitURL(l)
	rURL := config.ParseGitURL(r)

	if rURL != nil {
		return r, nil
	}
	if lURL == nil {
		return "", fmt.Errorf("fail to parse URL: %s", l)
	}
	if lURL.Repo == "" {
		lURL.Repo = r
	} else {
		lPath := lURL.Repo
		if !filepath.IsAbs(lPath) {
			lPath = "/" + lPath
		}
		lPath = filepath.Join(lPath, r)
		lPath = filepath.ToSlash(filepath.Clean(lPath))
		if !lURL.IsLocal() && len(lPath) > 0 && lPath[0] == '/' {
			lPath = lPath[1:]
		}
		lURL.Repo = lPath
	}

	return lURL.String(), nil
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
