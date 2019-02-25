package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestViperEnv(t *testing.T) {
	assert := assert.New(t)

	key := fmt.Sprintf("%s_%s", ViperEnvPrefix, "VERBOSE")
	os.Setenv(key, "2")
	assert.Equal(2, GetVerbose())

	key = fmt.Sprintf("%s_%s", ViperEnvPrefix, "LOGLEVEL")
	os.Setenv(key, "error")
	assert.Equal("error", GetLogLevel())

	key = fmt.Sprintf("%s_%s", ViperEnvPrefix, "SINGLE")
	os.Setenv(key, "true")
	assert.True(IsSingleMode())
	os.Setenv(key, "1")
	assert.True(IsSingleMode())
	os.Unsetenv(key)
	assert.False(IsSingleMode())
	os.Setenv(key, "0")
	assert.False(IsSingleMode())
	os.Setenv(key, "false")
	assert.False(IsSingleMode())
}
