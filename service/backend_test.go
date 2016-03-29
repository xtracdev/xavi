package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetConnectAddress(t *testing.T) {
	var testKVS = initKVStore(t)
	backend, err := buildBackend(testKVS, "hello-backend")
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
	_, err := buildBackend(testKVS, "no such backend")
	assert.NotNil(t, err)
}

func TestBuildTLSBackend(t *testing.T) {
	var testKVS = initKVStore(t)
	be, err := buildBackend(testKVS, "be-tls")
	if assert.Nil(t, err) {
		assert.Equal(t, true, be.TLSOnly)
		assert.NotNil(t, be.CACert)
	}
}

func TestBuildTLSBackendBadCert(t *testing.T) {
	var testKVS = initKVStore(t)
	_, err := buildBackend(testKVS, "be-tls-bogus-cert")
	if assert.NotNil(t, err) {
		t.Log(err.Error())
	}
}
