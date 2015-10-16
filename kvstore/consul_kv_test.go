package kvstore

import (
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func makeClient(t *testing.T) (*consulapi.Client, *testutil.TestServer) {
	return makeClientWithConfig(t)
}

func makeClientWithConfig(t *testing.T) (*consulapi.Client, *testutil.TestServer) {

	// Make client config
	conf := consulapi.DefaultConfig()

	// Create server
	server := testutil.NewTestServerConfig(t, nil)
	conf.Address = server.HTTPAddr

	// Create client
	client, err := consulapi.NewClient(conf)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	return client, server
}

func TestKVPutGetList(t *testing.T) {
	client, server := makeClient(t)
	defer server.Stop()
	consulKV := &ConsulKVStore{
		KV: client.KV(),
	}

	err := consulKV.Put("a", []byte("abc"))
	if assert.Nil(t, err) == false {
		println(err.Error())
	}

	b, err := consulKV.Get("a")
	assert.Nil(t, err)
	assert.Equal(t, []byte("abc"), b)

	kvp, err := consulKV.List("a")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(kvp))
}

func TestKVListError(t *testing.T) {
	client, server := makeClient(t)
	server.Stop()
	consulKV := &ConsulKVStore{
		KV: client.KV(),
	}

	_, err := consulKV.List("a")
	assert.NotNil(t, err)
}

func TestFlush(t *testing.T) {
	consulKV := &ConsulKVStore{}
	err := consulKV.Flush()
	assert.Nil(t, err)
}
