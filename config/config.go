package config

import (
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/kvstore"
)

//NullJSON represents the marshaling of a nil value
var (
	NullJSON = []byte("null")
)

func readKey(key string, kvs kvstore.KVStore) ([]byte, error) {
	log.Debug("Read key " + key)
	return kvs.Get(key)
}

//ListenContext is set to true if the listen command is being executed
var ListenContext bool
