package project

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// const for project
const (
	GroupDefault    = "default"
	GroupAll        = "all"
	GroupNotDefault = "notdefault"
)

func urlJoin(manifestURL, fetchURL, name string) (string, error) {
	var (
		u            *url.URL
		err          error
		manglePrefix = false
		mangleColumn = false
	)

	if strings.HasSuffix(manifestURL, "/") {
		manifestURL = strings.TrimRight(manifestURL, "/")
	}
	if strings.HasSuffix(manifestURL, ".git") {
		manifestURL = strings.TrimSuffix(manifestURL, ".git")
	}
	if !strings.Contains(manifestURL, "://") {
		slices := strings.SplitN(manifestURL, ":", 2)
		if len(slices) == 2 {
			manifestURL = strings.Join(slices, "/")
			mangleColumn = true
		}
		manifestURL = "gopher://" + manifestURL
		manglePrefix = true
	}
	u, err = url.Parse(manifestURL)
	if err != nil {
		return "", fmt.Errorf("bad manifest url - %s: %s", manifestURL, err)
	}
	u.Path = filepath.Clean(filepath.Join(u.Path, fetchURL, name))
	joinURL := u.String()

	if manglePrefix {
		joinURL = strings.TrimPrefix(joinURL, "gopher://")
		if mangleColumn {
			slices := strings.SplitN(joinURL, "/", 2)
			if len(slices) == 2 {
				joinURL = strings.Join(slices, ":")
			}
		}
	}
	return joinURL, nil
}

// MatchGroups checks if project has matched groups
func MatchGroups(match, groups string) bool {
	matchGroups := []string{}
	for _, g := range strings.Split(match, ",") {
		matchGroups = append(matchGroups, strings.TrimSpace(g))
	}
	if len(matchGroups) == 0 {
		matchGroups = append(matchGroups, GroupDefault)
	}

	projectGroups := []string{GroupAll}
	hasNotDefault := false
	for _, g := range strings.Split(groups, ",") {
		g = strings.TrimSpace(g)
		projectGroups = append(projectGroups, g)
		if g == GroupNotDefault {
			hasNotDefault = true
		}
	}
	if !hasNotDefault {
		projectGroups = append(projectGroups, GroupDefault)
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
