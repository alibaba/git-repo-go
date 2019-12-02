package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLoginFromEmail(t *testing.T) {
	assert := assert.New(t)

	input := []string{
		"User Name <user@example.com>",
		"User Name  <my-login@example.com>",
		"User Name<login@example.com>",
		"<login@example.com>",
		"login@example.com",
	}

	expect := []string{
		"user",
		"my-login",
		"login",
		"login",
		"login",
	}

	for i, email := range input {
		login := GetLoginFromEmail(email)
		assert.Equal(expect[i], login)
	}
}
