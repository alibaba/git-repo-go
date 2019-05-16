package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitURL(t *testing.T) {
	var (
		u       *GitURL
		address string
		assert  = assert.New(t)
	)

	for _, address = range []string{
		"http://git:pass@code.aone.alibaba-inc.com:8080/my/repos.git/",
		"http://git:pass@code.aone.alibaba-inc.com:8080/my/repos.git",
		"http://git:pass@code.aone.alibaba-inc.com:8080/my/repos",
		"http://git:pass@code.aone.alibaba-inc.com:8080/my/repos/",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("http://code.aone.alibaba-inc.com:8080", u.GetReviewURL())
			assert.Equal("http", u.Proto)
			assert.Equal("git:pass", u.User)
			assert.Equal("code.aone.alibaba-inc.com", u.Host)
			assert.Equal(8080, u.Port)
			assert.Equal("my/repos", u.Repo)
		}
	}

	for _, address = range []string{
		"https://code.aone.alibaba-inc.com/my/repos.git/",
		"https://code.aone.alibaba-inc.com/my/repos.git",
		"https://code.aone.alibaba-inc.com/",
		"https://code.aone.alibaba-inc.com",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("https://code.aone.alibaba-inc.com", u.GetReviewURL())
			assert.Equal("https", u.Proto)
			assert.Equal("code.aone.alibaba-inc.com", u.Host)
		}
	}

	for _, address = range []string{
		"ssh://git@code.aone.alibaba-inc.com:10022/my/repos.git/",
		"ssh://git@code.aone.alibaba-inc.com:10022/my/repos.git",
		"ssh://git@code.aone.alibaba-inc.com:10022/my/repos",
		"ssh://git@code.aone.alibaba-inc.com:10022/my/repos/",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("ssh://git@code.aone.alibaba-inc.com:10022", u.GetReviewURL())
			assert.Equal("ssh", u.Proto)
			assert.Equal("git", u.User)
			assert.Equal("code.aone.alibaba-inc.com", u.Host)
			assert.Equal(10022, u.Port)
			assert.Equal("my/repos", u.Repo)
		}
	}

	for _, address = range []string{
		"ssh://git@code.aone.alibaba-inc.com/my/repos.git/",
		"ssh://git@code.aone.alibaba-inc.com/my/repos.git",
		"ssh://git@code.aone.alibaba-inc.com/my/repos",
		"ssh://git@code.aone.alibaba-inc.com/my/repos/",
		"git@code.aone.alibaba-inc.com:my/repos.git/",
		"git@code.aone.alibaba-inc.com:my/repos.git",
		"git@code.aone.alibaba-inc.com:my/repos/",
		"git@code.aone.alibaba-inc.com:my/repos",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("ssh://git@code.aone.alibaba-inc.com", u.GetReviewURL())
			assert.Equal("ssh", u.Proto)
			assert.Equal("git", u.User)
			assert.Equal("code.aone.alibaba-inc.com", u.Host)
			assert.Equal("my/repos", u.Repo)
		}
	}

	for _, address = range []string{
		"ssh://code.aone.alibaba-inc.com/my/repos.git/",
		"ssh://code.aone.alibaba-inc.com/my/repos.git",
		"ssh://code.aone.alibaba-inc.com/my/repos",
		"ssh://code.aone.alibaba-inc.com/my/repos/",
		"code.aone.alibaba-inc.com:my/repos.git/",
		"code.aone.alibaba-inc.com:my/repos.git",
		"code.aone.alibaba-inc.com:my/repos/",
		"code.aone.alibaba-inc.com:my/repos",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("ssh://code.aone.alibaba-inc.com", u.GetReviewURL())
			assert.Equal("ssh", u.Proto)
			assert.Equal("code.aone.alibaba-inc.com", u.Host)
			assert.Equal("my/repos", u.Repo)
		}
	}

	for _, address = range []string{
		"ssh://git@code.aone.alibaba-inc.com:22/",
		"ssh://git@code.aone.alibaba-inc.com:22",
		"ssh://git@code.aone.alibaba-inc.com/",
		"ssh://git@code.aone.alibaba-inc.com",
		"git@code.aone.alibaba-inc.com:",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("ssh://git@code.aone.alibaba-inc.com", u.GetReviewURL())
			assert.Equal("ssh", u.Proto)
			assert.Equal("code.aone.alibaba-inc.com", u.Host)
		}
	}
}
