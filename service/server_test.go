package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildServerWithUnknownName(t *testing.T) {
	kvs := initKVStore(t)
	_, err := buildServer("no such server", kvs)
	assert.NotNil(t, err)
}
