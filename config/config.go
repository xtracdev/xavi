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
	log.Info("Read key " + key)
	return kvs.Get(key)
}
