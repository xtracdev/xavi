package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetConnectAddress(t *testing.T) {
	var testKVS = initKVStore(t)
	backend, err := buildBackend("hello-backend", testKVS)
	assert.Nil(t, err)
	assert.NotNil(t, backend)

	serverMap := make(map[string]string)
	for i := 0; i < 2; i++ {
		s, err := backend.getConnectAddress()
		assert.Nil(t, err)
		serverMap[s] = s
	}

	assert.Equal(t, 2, len(serverMap), "Expected all servers to be returned as a connect address")
}

func TestBuildBackendWithUnknownName(t *testing.T) {
	var testKVS = initKVStore(t)
	_, err := buildBackend("no such backend", testKVS)
	assert.NotNil(t, err)
}
