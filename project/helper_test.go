package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProjectUrlJoin(t *testing.T) {
	assert := assert.New(t)
	base := "https://github.com/jiangxin/manifest.git"

	u, err := urlJoin(base, ".", "my/repo")
	assert.Nil(err)
	assert.Equal("https://github.com/jiangxin/manifest/my/repo", u)

	u, err = urlJoin(base, "..", "my/repo")
	assert.Nil(err)
	assert.Equal("https://github.com/jiangxin/my/repo", u)

	u, err = urlJoin(base, "../..", "my/repo")
	assert.Nil(err)
	assert.Equal("https://github.com/my/repo", u)

	u, err = urlJoin(base, "../../../..", "my/repo")
	assert.Nil(err)
	assert.Equal("https://github.com/my/repo", u)

	base = "https://github.com/jiangxin/manifest.git/"

	u, err = urlJoin(base, ".", "my/repo")
	assert.Nil(err)
	assert.Equal("https://github.com/jiangxin/manifest/my/repo", u)

	u, err = urlJoin(base, "..", "my/repo")
	assert.Nil(err)
	assert.Equal("https://github.com/jiangxin/my/repo", u)

	base = "ssh://git@github.com/jiangxin/manifest.git/"

	u, err = urlJoin(base, "..", "my/repo")
	assert.Nil(err)
	assert.Equal("ssh://git@github.com/jiangxin/my/repo", u)

	base = "file:///root/manifest.git/"

	u, err = urlJoin(base, "..", "my/repo")
	assert.Nil(err)
	assert.Equal("file:///root/my/repo", u)

	u, err = urlJoin(base, "../../..", "my/repo")
	assert.Nil(err)
	assert.Equal("file:///my/repo", u)

	base = "git@github.com:jiangxin/manifest.git/"

	u, err = urlJoin(base, "..", "my/repo")
	assert.Nil(err)
	assert.Equal("git@github.com:jiangxin/my/repo", u)

	u, err = urlJoin(base, "../../..", "my/repo")
	assert.Nil(err)
	assert.Equal("git@github.com:my/repo", u)
}

func TestMatchGroups(t *testing.T) {
	var (
		match, groups string
	)

	assert := assert.New(t)

	match = ""
	groups = ""
	assert.True(MatchGroups(match, groups))

	match = "default"
	groups = ""
	assert.True(MatchGroups(match, groups))

	match = "all"
	groups = ""
	assert.True(MatchGroups(match, groups))

	match = "g3"
	groups = "g1,g2"
	assert.False(MatchGroups(match, groups))

	match = "-g1,g2"
	groups = "g1,g2"
	assert.True(MatchGroups(match, groups))

	match = "g1,-g2"
	groups = "g1,g2"
	assert.False(MatchGroups(match, groups))

	match = ""
	groups = "g1,notdefault"
	assert.False(MatchGroups(match, groups))

	match = "default"
	groups = "g1,notdefault"
	assert.False(MatchGroups(match, groups))

	match = "all"
	groups = "g1,notdefault"
	assert.True(MatchGroups(match, groups))

	match = "g1"
	groups = "g1,notdefault"
	assert.True(MatchGroups(match, groups))

	match = "g2"
	groups = "g1,notdefault"
	assert.False(MatchGroups(match, groups))
}
