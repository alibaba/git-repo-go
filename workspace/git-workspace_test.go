package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetReviewURL(t *testing.T) {
	var (
		assert = assert.New(t)
	)

	assert.Equal("http://example.com", getReviewURL("http://example.com/my/repos.git"))
	assert.Equal("http://example.com:8080", getReviewURL("http://user@example.com:8080/my/repos.git"))
	assert.Equal("https://example.com:8080", getReviewURL("https://user@example.com:8080"))
	assert.Equal("example.com", getReviewURL("ssh://example.com/my/repos.git"))
	assert.Equal("example.com", getReviewURL("ssh://git@example.com:29418/my/repos.git"))
	assert.Equal("example.com", getReviewURL("git@example.com:my/repos.git"))
}
