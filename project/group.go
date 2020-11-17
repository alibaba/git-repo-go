package project

import (
	"strings"
)

const (
	groupDefaultConst    = "default"
	groupAllConst        = "all"
	groupNotDefaultConst = "notdefault"
)

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
