package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMin(t *testing.T) {
	assert := assert.New(t)

	m := min(200, 124, 200, 89, 321)
	assert.Equal(89, int(m))

	m = min(200)
	assert.Equal(200, int(m))
}
