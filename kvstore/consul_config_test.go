package kvstore

import (
	"github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

//TODO - do we need consulConfigFromEnv anymore?
func TestDefault(t *testing.T) {
	u, err := url.Parse("consul://127.0.0.1:8500")
	assert.Nil(t, err)
	config, err := consulConfigFromEnv(u)
	assert.Equal(t, nil, err)
	assert.Equal(t, "127.0.0.1:8500", config.Address)
}

func TestEnvConfig(t *testing.T) {
	u, err := url.Parse("consul://127.0.1.1:9876")
	assert.Nil(t, err)
	config, err := consulConfigFromEnv(u)
	assert.Equal(t, nil, err)
	assert.Equal(t, "127.0.1.1:9876", config.Address)
}
