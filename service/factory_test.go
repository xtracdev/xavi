package service

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"github.com/xtracdev/xavi/plugin"
	"github.com/xtracdev/xavi/plugin/logging"
)

func initKVStore(t *testing.T) kvstore.KVStore {
	return config.BuildKVStoreTestConfig(t)
}

func TestServerFactory(t *testing.T) {
	plugin.RegisterWrapperFactory("Logging", logging.NewLoggingWrapper)

	var testKVS = initKVStore(t)
	service, err := BuildServiceForListener("listener", "0.0.0.0:8000", testKVS)
	assert.Nil(t, err)
	s := fmt.Sprintf("%s", service)
	println(s)
	assert.True(t, strings.Contains(s, "route1"))
	assert.True(t, strings.Contains(s, "hello-backend"))
	assert.True(t, strings.Contains(s, "listener"))
	assert.True(t, strings.Contains(s, "0.0.0.0:8000"))
}

func TestServerFactoryWithUnknownListener(t *testing.T) {
	var testKVS = initKVStore(t)
	_, err := BuildServiceForListener("no such listener", "0.0.0.0:8000", testKVS)
	assert.NotNil(t, err)
}
