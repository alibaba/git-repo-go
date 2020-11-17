package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
