package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareVersion(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(0, CompareVersion("10", "10"))
	assert.Equal(1, CompareVersion("11", "10"))
	assert.Equal(-1, CompareVersion("10", "11"))
	assert.Equal(-1, CompareVersion("2.2.3", "2.10.1"))
	assert.Equal(-1, CompareVersion("1.2.3", "1.2.3.4"))
	assert.Equal(1, CompareVersion("1.2.3.4", "1.2.3"))
	assert.Equal(1, CompareVersion("1.2.3", "1.2.3.rc1"))
	assert.Equal(-1, CompareVersion("1.2.3.rc1", "1.2.3"))
	assert.Equal(1, CompareVersion("0.1.0", "undefined"))
	assert.Equal(-1, CompareVersion("undefined", "0.1.0"))
}
