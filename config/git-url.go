package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	// GitHTTPProtocolPattern indicates git over HTTP protocol
	GitHTTPProtocolPattern = regexp.MustCompile(`^(?P<proto>http|https)://((?P<user>.*?)@)?(?P<host>[^/]+?)(:(?P<port>[0-9]+))?(/(?P<repo>.*?)(\.git)?/?)?$`)
	// GitSSHProtocolPattern indicates git over SSH protocol
	GitSSHProtocolPattern = regexp.MustCompile(`^(?P<proto>ssh)://((?P<user>.*?)@)?(?P<host>[^/]+?)(:(?P<port>[0-9]+))?(/(?P<repo>.+?)(\.git)?)?/?$`)
	// GitRsyncProtocolPattern indicates rsync style git over SSH protocol
	GitRsyncProtocolPattern = regexp.MustCompile(`^((?P<user>.*?)@)?(?P<host>[^/:]+?):(?P<repo>.*?)(\.git)?/?$`)
)

// GitURL holds Git URL
type GitURL struct {
	Proto string
	User  string
	Host  string
	Port  int
	Repo  string
}

// GetReviewURL returns review URL
func (v GitURL) GetReviewURL() string {
	var u string
	if v.Proto == "http" || v.Proto == "https" {
		u = v.Proto + "://"
		u += v.Host
		if v.Port > 0 && v.Port != 80 && v.Port != 443 {
			u += fmt.Sprintf(":%d", v.Port)
		}
	} else if v.Proto == "ssh" {
		u = v.Proto + "://"
		if v.User != "" {
			u += v.User + "@"
		}
		u += v.Host
		if v.Port > 0 && v.Port != 22 {
			u += fmt.Sprintf(":%d", v.Port)
		}
	} else {
		u = v.Host
	}
	return u
}

func getMatchedGitURL(re *regexp.Regexp, data string) *GitURL {
	matches := re.FindStringSubmatch(data)
	if len(matches) == 0 {
		return nil
	}
	gitURL := GitURL{
		Proto: "ssh",
	}
	for i, name := range re.SubexpNames() {
		if name == "" {
			continue
		}
		switch name {
		case "proto":
			gitURL.Proto = matches[i]
		case "user":
			gitURL.User = matches[i]
		case "host":
			gitURL.Host = matches[i]
		case "port":
			port, err := strconv.Atoi(matches[i])
			if err == nil {
				gitURL.Port = port
			}
		case "repo":
			gitURL.Repo = matches[i]
		}
	}

	return &gitURL
}

// ParseGitURL parses address and returns GitURL
func ParseGitURL(address string) *GitURL {
	var (
		gitURL *GitURL
	)

	if strings.Contains(address, "://") {
		proto := strings.Split(address, "://")[0]
		if proto != "http" && proto != "https" && proto != "ssh" {
			return nil
		}
	}

	gitURL = getMatchedGitURL(GitHTTPProtocolPattern, address)
	if gitURL != nil {
		return gitURL
	}
	gitURL = getMatchedGitURL(GitSSHProtocolPattern, address)
	if gitURL != nil {
		return gitURL
	}
	gitURL = getMatchedGitURL(GitRsyncProtocolPattern, address)
	if gitURL != nil {
		return gitURL
	}
	return nil
}
