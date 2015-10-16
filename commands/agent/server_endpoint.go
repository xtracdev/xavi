package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"io/ioutil"
	"net/http"
)

//Each REST resource should defined its base URI here
const (
	serversURI = "/v1/servers/"
)

//Errors
var (
	errServerNotFound         = errors.New("Server definition not found")
	errServiceResourceMissing = errors.New("Server resource not present in url - expected /v1/servers/server-resource")
)

//Exported ServerDef for external reference
var ServerDefCmd ServerDef

//ServerDef is used to hang the ApiCommand functions needed for exposing server def capabilities
//via a REST API
type ServerDef struct{}

//GetURIRoot returns the URI root used to serve server def API calls
func (ServerDef) GetURIRoot() string {
	return serversURI
}

//PutDefinition creates or updates a server definition
func (ServerDef) PutDefinition(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	body, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	log.Info(fmt.Sprintf("Put request with payload %s", string(body)))
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, nil
	}

	serverName := resourceIDFromURI(req.RequestURI)
	log.Info(serverName)
	if serverName == "" {
		resp.WriteHeader(http.StatusNotFound)
		return nil, errServiceResourceMissing
	}

	serverConfig := new(config.ServerConfig)
	err = json.Unmarshal(body, serverConfig)
	if err != nil {
		log.Warn("Error unmarshaling request body")
		resp.WriteHeader(http.StatusBadRequest)
		return nil, err
	}

	serverConfig.Name = serverName
	err = serverConfig.Store(kvs)
	if err != nil {
		log.Warn("Error persisting server definition")
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	err = kvs.Flush()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

//GetDefinitionList returns a list of server definitions
func (ServerDef) GetDefinitionList(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {

	servers, err := config.ListServerConfigs(kvs)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	if servers == nil {
		servers = make([]*config.ServerConfig, 0)
	}

	return servers, nil

}

//GetDefinition retrieves a specific server definition
func (ServerDef) GetDefinition(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	serverName := resourceIDFromURI(req.RequestURI)

	serverConfig, err := config.ReadServerConfig(serverName, kvs)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	if serverConfig == nil {
		resp.WriteHeader(http.StatusNotFound)
		return nil, errServerNotFound
	}

	return serverConfig, err

}

func (ServerDef) DoPost(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	resp.WriteHeader(http.StatusMethodNotAllowed)
	return nil, nil
}
