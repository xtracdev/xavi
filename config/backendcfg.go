package config

import (
	"encoding/json"
	"fmt"
	log "github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/kvstore"
)

//BackendConfig defines the data stored for a Backend definition
type BackendConfig struct {
	Name               string
	ServerNames        []string
	LoadBalancerPolicy string
}

//JSONToBackend unmarshals a JSON representation of a BackendCOnfig
func JSONToBackend(bytes []byte) *BackendConfig {
	var b *BackendConfig
	if bytes == nil {
		return b
	}

	b = new(BackendConfig)
	json.Unmarshal(bytes, b)
	return b
}

//Store persists a backend definition using the supplied key value store
func (backendConfig *BackendConfig) Store(kvs kvstore.KVStore) error {
	b, err := json.Marshal(backendConfig)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("backends/%s", backendConfig.Name)
	log.Info(fmt.Sprintf("Adding %s under key %s", string(b), key))
	err = kvs.Put(key, b)
	if err != nil {
		return err
	}

	return nil
}

//ReadBackendConfig reads a backend config defnition using the supplied KVS
func ReadBackendConfig(name string, kvs kvstore.KVStore) (*BackendConfig, error) {
	//Read the definition from the key store
	bv, err := readKey("backends/"+name, kvs)
	if err != nil {
		return nil, err
	}

	return JSONToBackend(bv), nil
}

//ListBackendConfigs lists the backend definitions present in the supplied KVS
func ListBackendConfigs(kvs kvstore.KVStore) ([]*BackendConfig, error) {
	pairs, err := kvs.List("backends/")
	if err != nil {
		return nil, err
	}

	var backends []*BackendConfig
	for _, p := range pairs {
		backends = append(backends, JSONToBackend(p.Value))
	}

	return backends, nil
}
