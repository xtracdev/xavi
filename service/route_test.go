package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildRouteWithUnknownName(t *testing.T) {
	kvs := initKVStore(t)
	_, err := buildRoute("no such route", kvs)
	assert.NotNil(t, err)
}
