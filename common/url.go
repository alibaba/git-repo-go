package common

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alibaba/git-repo-go/config"
)

// URLJoin appends fetch path (in remote element) and project name to manifest url.
func URLJoin(u string, paths ...string) (string, error) {
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
