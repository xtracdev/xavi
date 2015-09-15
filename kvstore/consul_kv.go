package kvstore

import (
	"fmt"
	log "github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	consulapi "github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/hashicorp/consul/api"
	"net/url"
)

//ConsulKVStore is an implementation of KVStore using Consul.
type ConsulKVStore struct {
	KV *consulapi.KV
}

//NewConsulKVStore creates a new instance of the ConsulKVStore
func NewConsulKVStore(u *url.URL) (*ConsulKVStore, error) {
	consulConfig, err := consulConfigFromEnv(u)
	if err != nil {
		return nil, err
	}

	client, err := consulapi.NewClient(consulConfig)
	if err != nil {
		return nil, err
	}

	return &ConsulKVStore{
		KV: client.KV(),
	}, nil
}

//Put stores a value using the given key in consul
func (kvs *ConsulKVStore) Put(key string, value []byte) error {
	p := &consulapi.KVPair{Key: key, Value: value}
	_, err := kvs.KV.Put(p, nil)
	return err
}

//Get returns the value, if any, stored under the given key. A
//nil value is returned if not present in the KVS
func (kvs *ConsulKVStore) Get(key string) ([]byte, error) {
	log.Info(fmt.Sprintf("Retrieving %s from consul store", key))
	kvPair, _, err := kvs.KV.Get(key, nil)
	if err != nil {
		return nil, err
	}

	if kvPair != nil {
		return kvPair.Value, nil
	}
	return nil, nil

}

//List returns a list of the objects stored under the given key. Nil
//is returned if nothing is present for the given key
func (kvs *ConsulKVStore) List(key string) ([]*KVPair, error) {
	log.Info("List values under ", key)
	pairs, _, err := kvs.KV.List(key, nil)
	if err != nil {
		log.Info("Error listing keys: ", err.Error())
		return nil, err
	}

	var kvpairs []*KVPair
	for _, v := range pairs {
		kvpairs = append(kvpairs, &KVPair{v.Key, v.Value})
	}

	return kvpairs, nil
}

//Flush is a no-op for consul backed KVStores
func (kvs *ConsulKVStore) Flush() error {
	return nil
}
