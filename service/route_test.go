package service

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestBuildRouteWithUnknownName(t *testing.T) {
	kvs := initKVStore(t)
	_, err := buildRoute("no such route", kvs)
	assert.NotNil(t, err)
}

func TestBuildRoutesWithNoPlugin(t *testing.T) {
	kvs := initKVStore(t)
	_, err := buildRoute("r1", kvs)
	if assert.NotNil(t, err) {
		assert.True(t, strings.Contains(err.Error(), "MultiRoute"))
	}

}
