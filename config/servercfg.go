package config

import (
	"encoding/json"
	"fmt"
	log "github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/kvstore"
)

//ServerConfig represents a server configuration definition
type ServerConfig struct {
	Name                string
	Address             string
	Port                int
	PingURI             string
	HealthCheck         string
	HealthCheckInterval int //In milliseconds
	HealthCheckTimeout  int //In milliseconds
}

//JSONToServer unmarshals a JSON representation of a server definition
func JSONToServer(bytes []byte) *ServerConfig {
	var s *ServerConfig
	if bytes == nil {
		return s
	}

	s = new(ServerConfig)
	json.Unmarshal(bytes, s)
	return s
}

//Store persists the server configuration definition in the supplied KVS
func (serverConfig *ServerConfig) Store(kvs kvstore.KVStore) (err error) {
	b, err := json.Marshal(serverConfig)
	if err != nil {
		return
	}

	key := fmt.Sprintf("servers/%s", serverConfig.Name)
	log.Info(fmt.Sprintf("adding %s under key %s", string(b), key))
	err = kvs.Put(key, b)
	return
}

//ReadServerConfig retrieves the named server defnition using the supplied KVS
func ReadServerConfig(name string, kvs kvstore.KVStore) (*ServerConfig, error) {

	//Read the definition from the key store
	bv, err := readKey("servers/"+name, kvs)
	if err != nil {
		return nil, err
	}

	return JSONToServer(bv), nil
}

//ListServerConfigs returns a list of the server configurations
//present in the supplied KVS
func ListServerConfigs(kvs kvstore.KVStore) ([]*ServerConfig, error) {
	pairs, err := kvs.List("servers/")
	if err != nil {
		return nil, err
	}

	var servers []*ServerConfig
	for _, p := range pairs {
		servers = append(servers, JSONToServer(p.Value))
	}

	return servers, nil
}
