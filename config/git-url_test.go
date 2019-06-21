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
		"http://git:pass@code.aone.alibaba-inc.com:8080/my/repo.git/",
		"http://git:pass@code.aone.alibaba-inc.com:8080/my/repo.git",
		"http://git:pass@code.aone.alibaba-inc.com:8080/my/repo",
		"http://git:pass@code.aone.alibaba-inc.com:8080/my/repo/",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("http://code.aone.alibaba-inc.com:8080", u.GetReviewURL())
			assert.Equal("http", u.Proto)
			assert.Equal("git:pass", u.User)
			assert.Equal("code.aone.alibaba-inc.com", u.Host)
			assert.Equal(8080, u.Port)
			assert.Equal("my/repo", u.Repo)
		}
	}

	for _, address = range []string{
		"https://code.aone.alibaba-inc.com/my/repo.git/",
		"https://code.aone.alibaba-inc.com/my/repo.git",
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
		"ssh://git@code.aone.alibaba-inc.com:10022/my/repo.git/",
		"ssh://git@code.aone.alibaba-inc.com:10022/my/repo.git",
		"ssh://git@code.aone.alibaba-inc.com:10022/my/repo",
		"ssh://git@code.aone.alibaba-inc.com:10022/my/repo/",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("ssh://git@code.aone.alibaba-inc.com:10022", u.GetReviewURL())
			assert.Equal("ssh", u.Proto)
			assert.Equal("git", u.User)
			assert.Equal("code.aone.alibaba-inc.com", u.Host)
			assert.Equal(10022, u.Port)
			assert.Equal("my/repo", u.Repo)
		}
	}

	for _, address = range []string{
		"ssh://git@code.aone.alibaba-inc.com/my/repo.git/",
		"ssh://git@code.aone.alibaba-inc.com/my/repo.git",
		"ssh://git@code.aone.alibaba-inc.com/my/repo",
		"ssh://git@code.aone.alibaba-inc.com/my/repo/",
		"git@code.aone.alibaba-inc.com:my/repo.git/",
		"git@code.aone.alibaba-inc.com:my/repo.git",
		"git@code.aone.alibaba-inc.com:my/repo/",
		"git@code.aone.alibaba-inc.com:my/repo",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("ssh://git@code.aone.alibaba-inc.com", u.GetReviewURL())
			assert.Equal("ssh", u.Proto)
			assert.Equal("git", u.User)
			assert.Equal("code.aone.alibaba-inc.com", u.Host)
			assert.Equal("my/repo", u.Repo)
		}
	}

	for _, address = range []string{
		"ssh://code.aone.alibaba-inc.com/my/repo.git/",
		"ssh://code.aone.alibaba-inc.com/my/repo.git",
		"ssh://code.aone.alibaba-inc.com/my/repo",
		"ssh://code.aone.alibaba-inc.com/my/repo/",
		"code.aone.alibaba-inc.com:my/repo.git/",
		"code.aone.alibaba-inc.com:my/repo.git",
		"code.aone.alibaba-inc.com:my/repo/",
		"code.aone.alibaba-inc.com:my/repo",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("ssh://code.aone.alibaba-inc.com", u.GetReviewURL())
			assert.Equal("ssh", u.Proto)
			assert.Equal("code.aone.alibaba-inc.com", u.Host)
			assert.Equal("my/repo", u.Repo)
		}
	}

	for _, address = range []string{
		"ssh://git@code.aone.alibaba-inc.com:22/",
		"ssh://git@code.aone.alibaba-inc.com:22",
		"ssh://git@code.aone.alibaba-inc.com/",
		"ssh://git@code.aone.alibaba-inc.com",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("ssh://git@code.aone.alibaba-inc.com", u.GetReviewURL())
			assert.Equal("ssh", u.Proto)
			assert.Equal("code.aone.alibaba-inc.com", u.Host)
		}
	}

	for _, address = range []string{
		"git://code.aone.alibaba-inc.com/my/repo.git/",
		"git://code.aone.alibaba-inc.com/my/repo.git",
		"git://code.aone.alibaba-inc.com/",
		"git://code.aone.alibaba-inc.com",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("code.aone.alibaba-inc.com", u.GetReviewURL())
			assert.Equal("git", u.Proto)
			assert.Equal("code.aone.alibaba-inc.com", u.Host)
		}
	}

	for _, address = range []string{
		"file:///path/of/repo.git/",
		"file:///path/of/repo.git",
		"/path/of/repo.git",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("", u.GetReviewURL())
			assert.Equal("file", u.Proto)
			assert.Equal("/path/of/repo.git", u.Repo)
		}
	}

	for _, address = range []string{
		"ftp://host/path/of/repo.git/",
		"ftp://host/path/of/repo.git",
	} {
		u = ParseGitURL(address)
		assert.Nil(u)
	}
}
