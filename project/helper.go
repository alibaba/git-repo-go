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

func urlJoin(u string, names ...string) (string, error) {
	if len(names) == 0 {
		return u, nil
	}

	// If names[0] is an URL, ignore u and start with name[0]
	if strings.Contains(names[0], ":") {
		return _urlJoin(names...)
	}

	// Remove last part of manifest url
	paths := []string{u, ".."}
	paths = append(paths, names...)
	return _urlJoin(paths...)
}

func _urlJoin(names ...string) (string, error) {
	var (
		u            *url.URL
		err          error
		manglePrefix = false
		mangleColumn = false
	)

	if len(names) == 0 {
		return "", nil
	} else if len(names) == 1 {
		return names[0], nil
	}

	for strings.HasSuffix(names[0], "/") {
		names[0] = strings.TrimRight(names[0], "/")
	}
	if names[0] == "" {
		names[0] = "/"
	}

	// names[1] is an URL
	if strings.Contains(names[1], ":") {
		return _urlJoin(names[1:]...)
	}

	if !strings.Contains(names[0], "://") {
		slices := strings.SplitN(names[0], ":", 2)
		if len(slices) == 2 {
			names[0] = strings.Join(slices, "/")
			mangleColumn = true
		}
		names[0] = "gopher://" + names[0]
		manglePrefix = true
	}
	u, err = url.Parse(names[0])
	if err != nil {
		return "", fmt.Errorf("bad manifest url - %s: %s", names[0], err)
	}

	ps := []string{u.Path}
	ps = append(ps, names[1:]...)
	u.Path = filepath.Clean(filepath.Join(ps...))
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
