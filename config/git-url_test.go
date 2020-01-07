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
		"http://git:pass@example.com:8080/my/repo.git/",
		"http://git:pass@example.com:8080/my/repo.git",
		"http://git:pass@example.com:8080/my/repo",
		"http://git:pass@example.com:8080/my/repo/",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("http://example.com:8080", u.GetRootURL())
			assert.Equal("http", u.Proto)
			assert.Equal("git:pass", u.User)
			assert.Equal("example.com", u.Host)
			assert.Equal(8080, u.Port)
			assert.Equal("my/repo", u.Repo)
		}
	}

	for _, address = range []string{
		"https://example.com/my/repo.git/",
		"https://example.com/my/repo.git",
		"https://example.com/",
		"https://example.com",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("https://example.com", u.GetRootURL())
			assert.Equal("https", u.Proto)
			assert.Equal("example.com", u.Host)
		}
	}

	for _, address = range []string{
		"ssh://git@example.com:10022/my/repo.git/",
		"ssh://git@example.com:10022/my/repo.git",
		"ssh://git@example.com:10022/my/repo",
		"ssh://git@example.com:10022/my/repo/",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("ssh://git@example.com:10022", u.GetRootURL())
			assert.Equal("ssh", u.Proto)
			assert.Equal("git", u.User)
			assert.Equal("example.com", u.Host)
			assert.Equal(10022, u.Port)
			assert.Equal("my/repo", u.Repo)
		}
	}

	for _, address = range []string{
		"ssh://git@example.com/my/repo.git/",
		"ssh://git@example.com/my/repo.git",
		"ssh://git@example.com/my/repo",
		"ssh://git@example.com/my/repo/",
		"git@example.com:my/repo.git/",
		"git@example.com:my/repo.git",
		"git@example.com:my/repo/",
		"git@example.com:my/repo",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("ssh://git@example.com", u.GetRootURL())
			assert.Equal("ssh", u.Proto)
			assert.Equal("git", u.User)
			assert.Equal("example.com", u.Host)
			assert.Equal("my/repo", u.Repo)
		}
	}

	for _, address = range []string{
		"ssh://example.com/my/repo.git/",
		"ssh://example.com/my/repo.git",
		"ssh://example.com/my/repo",
		"ssh://example.com/my/repo/",
		"example.com:my/repo.git/",
		"example.com:my/repo.git",
		"example.com:my/repo/",
		"example.com:my/repo",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("ssh://example.com", u.GetRootURL())
			assert.Equal("ssh", u.Proto)
			assert.Equal("example.com", u.Host)
			assert.Equal("my/repo", u.Repo)
		}
	}

	for _, address = range []string{
		"ssh://git@example.com:22/",
		"ssh://git@example.com:22",
		"ssh://git@example.com/",
		"ssh://git@example.com",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("ssh://git@example.com", u.GetRootURL())
			assert.Equal("ssh", u.Proto)
			assert.Equal("example.com", u.Host)
		}
	}

	for _, address = range []string{
		"git://example.com/my/repo.git/",
		"git://example.com/my/repo.git",
		"git://example.com/",
		"git://example.com",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("example.com", u.GetRootURL())
			assert.Equal("git", u.Proto)
			assert.Equal("example.com", u.Host)
		}
	}

	for _, address = range []string{
		"file:///path/of/repo.git/",
		"file:///path/of/repo.git",
		"/path/of/repo.git",
	} {
		u = ParseGitURL(address)
		if assert.NotNil(u) {
			assert.Equal("", u.GetRootURL())
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

func TestGitURLWithoutRepo(t *testing.T) {
	var (
		u      *GitURL
		assert = assert.New(t)
	)

	u = ParseGitURL("http://example.com")
	if assert.NotNil(u) {
		assert.Equal("http", u.Proto)
		assert.Equal("example.com", u.Host)
		assert.Equal(0, u.Port)
		assert.Equal("", u.User)
	}

	u = ParseGitURL("https://example.com/")
	if assert.NotNil(u) {
		assert.Equal("https", u.Proto)
		assert.Equal("example.com", u.Host)
		assert.Equal(0, u.Port)
		assert.Equal("", u.User)
	}

	u = ParseGitURL("http://example.com:8080/")
	if assert.NotNil(u) {
		assert.Equal("http", u.Proto)
		assert.Equal("example.com", u.Host)
		assert.Equal(8080, u.Port)
		assert.Equal("", u.User)
	}

	u = ParseGitURL("https://user:pass@example.com:1234/")
	if assert.NotNil(u) {
		assert.Equal("https", u.Proto)
		assert.Equal("example.com", u.Host)
		assert.Equal(1234, u.Port)
		assert.Equal("user:pass", u.User)
	}

	u = ParseGitURL("ssh://example.com")
	if assert.NotNil(u) {
		assert.Equal("ssh", u.Proto)
		assert.Equal("example.com", u.Host)
		assert.Equal(0, u.Port)
		assert.Equal("", u.User)
	}

	u = ParseGitURL("ssh://user@example.com:29418/")
	if assert.NotNil(u) {
		assert.Equal("ssh", u.Proto)
		assert.Equal("example.com", u.Host)
		assert.Equal(29418, u.Port)
		assert.Equal("user", u.User)
	}

	u = ParseGitURL("user@example.com")
	assert.Nil(u)
}
