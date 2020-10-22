package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetReviewRef(t *testing.T) {
	var (
		ref string
		err error
	)

	assert := assert.New(t)

	sshInfo := SSHInfo{}
	ref, err = sshInfo.GetReviewRef("123", "1")
	assert.NotNil(err)
	assert.Empty(ref)

	sshInfo.ReviewRefPattern = "refs/heads/master"
	ref, err = sshInfo.GetReviewRef("123", "1")
	assert.Nil(err)
	assert.Equal("refs/heads/master", ref)

	sshInfo.ReviewRefPattern = "refs/changes/{id:right:2}/{id}/{patch}"
	ref, err = sshInfo.GetReviewRef("123", "1")
	assert.Nil(err)
	assert.Equal("refs/changes/23/123/1", ref)

	sshInfo.ReviewRefPattern = "refs/changes/{id:right:2}/{id}/{patch}"
	ref, err = sshInfo.GetReviewRef("9", "1")
	assert.Nil(err)
	assert.Equal("refs/changes/09/9/1", ref)

	sshInfo.ReviewRefPattern = "refs/changes/{id:left:2}/{id}/{patch}"
	ref, err = sshInfo.GetReviewRef("123", "1")
	assert.Nil(err)
	assert.Equal("refs/changes/12/123/1", ref)

	sshInfo.ReviewRefPattern = "refs/changes/{id:left:2}/{id}/{patch}"
	ref, err = sshInfo.GetReviewRef("9", "1")
	assert.Nil(err)
	assert.Equal("refs/changes/09/9/1", ref)

	sshInfo.ReviewRefPattern = "refs/pull/{id}/head"
	ref, err = sshInfo.GetReviewRef("123", "1")
	assert.Nil(err)
	assert.Equal("refs/pull/123/head", ref)

	sshInfo.ReviewRefPattern = "refs}/pull/{id}/head"
	ref, err = sshInfo.GetReviewRef("123", "1")
	assert.Nil(err)
	assert.Equal("refs}/pull/123/head", ref, "unmatched right quote")

	sshInfo.ReviewRefPattern = "refs/{pull}/{id}/head"
	ref, err = sshInfo.GetReviewRef("123", "1")
	assert.Nil(err)
	assert.Equal("refs/{pull}/123/head", ref, "unknown key")
}
