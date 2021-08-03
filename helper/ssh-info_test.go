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

func TestSSHInfoFromString(t *testing.T) {
	assert := assert.New(t)

	for _, data := range []struct {
		SSHInfo string
		ErrMsg  string
	}{
		{
			SSHInfo: "",
			ErrMsg:  "empty ssh_info",
		},
		{
			SSHInfo: `
		<!DOCTYPE HTML PUBLIC "-//IETF//DTD HTML 2.0//EN">
		<html>
		<head><title>302 Found</title></head>
		<body bgcolor="white">
		<h1>302 Found</h1>
		<p>The requested resource resides temporarily under a different URI.</p>
		<hr/>Powered by Tengine</body>
		</html>`,
			ErrMsg: "ssh_info returns a normal HTML response",
		},
		{
			SSHInfo: `NOT_AVAILABLE`,
			ErrMsg:  "",
		},
		{
			SSHInfo: `host.name 22`,
			ErrMsg:  "",
		},
		{
			SSHInfo: `bad host-name 22`,
			ErrMsg:  "invalid ssh_info response: invalid character 'b' looking for beginning of value",
		},
		{
			SSHInfo: `{"host":"codeup.aliyun.com","port":22,"type":"agit"}`,
			ErrMsg:  "",
		},
		{
			SSHInfo: `{
					"host":"codeup.aliyun.com",
					"port":22,"type":"agit"
				  }`,
			ErrMsg: "",
		},
		{
			SSHInfo: `{"host":"codeup.aliyun.com 22","port":22,"type":"agit"}`,
			ErrMsg:  "ssh_info has invalid host name: codeup.aliyun.com 22",
		},
		{
			SSHInfo: `{
					"host":"codeup.aliyun.com",
					"port":22,
					"user":"!me",
					"type":"agit"
				   }`,
			ErrMsg: "ssh_info has invalid user name: !me",
		},
	} {
		sshInfo, err := sshInfoFromString(data.SSHInfo)
		if data.ErrMsg == "" {
			assert.NotNil(sshInfo, "nil ssh_info for: %s", data.SSHInfo)
			assert.Nil(err, "error is not nil for: %s", data.SSHInfo)
		} else {
			assert.Nil(sshInfo, "ssh_info is not nil for: %s", data.SSHInfo)
			assert.NotNil(err, "error is nil for: %s", data.SSHInfo)
			assert.Equal(data.ErrMsg, err.Error(), "unmatch error msg for: %s", data.SSHInfo)
		}
	}
}
