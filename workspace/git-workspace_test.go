package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetReviewURL(t *testing.T) {
	var (
		r      string
		err    error
		assert = assert.New(t)
	)

	r, err = getReviewURL("http://example.com/my/repos.git")
	assert.Equal("http://example.com", r)
	assert.Nil(err)

	r, err = getReviewURL("http://user@example.com:8080/my/repos.git")
	assert.Equal("http://example.com:8080", r)
	assert.Nil(err)

	r, err = getReviewURL("https://user@example.com:8080")
	assert.Equal("https://example.com:8080", r)
	assert.Nil(err)

	r, err = getReviewURL("ssh://example.com/my/repos.git")
	assert.Equal("example.com", r)
	assert.Nil(err)

	r, err = getReviewURL("ssh://git@example.com:29418/my/repos.git")
	assert.Equal("example.com", r)
	assert.Nil(err)

	r, err = getReviewURL("git@example.com:my/repos.git")
	assert.Equal("example.com", r)
	assert.Nil(err)

	r, err = getReviewURL("file:///workspace/my/repos.git")
	assert.Equal("", r)
	assert.NotNil(err)

	r, err = getReviewURL("ftp:///workspace/my/repos.git")
	assert.Equal("", r)
	assert.NotNil(err)
}
