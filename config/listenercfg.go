package config

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/kvstore"
)

//ListenerConfig represents a listener configuration definition
type ListenerConfig struct {
	Name           string
	RouteNames     []string
	HealthEndpoint bool
}

//JSONToListener unmarshals the JSON representation of a listener definition
func JSONToListener(bytes []byte) *ListenerConfig {
	var l *ListenerConfig
	if bytes == nil {
		return l
	}

	l = new(ListenerConfig)
	if err := json.Unmarshal(bytes, l); err != nil {
		log.Warn("Error unmarshalling ListenerConfig:", err.Error())
		l = nil
	}
	return l
}

//Store persists the listener defintions in the supplied key value store
func (listenerConfig *ListenerConfig) Store(kvs kvstore.KVStore) (err error) {
	b, err := json.Marshal(listenerConfig)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("listeners/%s", listenerConfig.Name)
	log.Info(fmt.Sprintf("adding %s under key %s", string(b), key))
	err = kvs.Put(key, b)
	return
}

//ReadListenerConfig retrieves the named listener definition from the supplied KVS
func ReadListenerConfig(name string, kvs kvstore.KVStore) (*ListenerConfig, error) {

	//Read the definition from the key store
	bv, err := readKey("listeners/"+name, kvs)
	if err != nil {
		return nil, err
	}

	return JSONToListener(bv), nil
}

//ListListenerConfigs retrieves the listener defs from the KVS
func ListListenerConfigs(kvs kvstore.KVStore) ([]*ListenerConfig, error) {
	pairs, err := kvs.List("listeners/")
	if err != nil {
		return nil, err
	}

	var listeners []*ListenerConfig
	for _, p := range pairs {
		listeners = append(listeners, JSONToListener(p.Value))
	}

	return listeners, nil
}
