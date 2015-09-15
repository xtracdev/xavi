package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/config"
	"github.com/xtracdev/xavi/kvstore"
	"io/ioutil"
	"net/http"
)

const (
	backendsURI = "/v1/backends/"
)

//Errors
var (
	errBackendNotFound        = errors.New("Backend definition not found")
	errBackendResourceMissing = errors.New("Backend resource not present in url - expected /v1/backends/backend-resource")
)

//Exported BackendDef for external reference
var BackendDefCmd BackendDef

//BackendDef is used to hang the ApiCommand functions needed for exposing backend def capabilities
//via a REST API
type BackendDef struct{}

//GetURIRoot returns the URI root used to serve backend def API calls
func (BackendDef) GetURIRoot() string {
	return backendsURI
}

//PutDefinition creates or updates a backend definition
func (BackendDef) PutDefinition(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	body, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	log.Info(fmt.Sprintf("Put request with payload %s", string(body)))
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, nil
	}

	backendName := resourceIDFromURI(req.RequestURI)
	log.Info(backendName)
	if backendName == "" {
		resp.WriteHeader(http.StatusNotFound)
		return nil, errServiceResourceMissing
	}

	backendConfig := new(config.BackendConfig)
	err = json.Unmarshal(body, backendConfig)
	if err != nil {
		log.Warn("Error unmarshaling request body")
		resp.WriteHeader(http.StatusBadRequest)
		return nil, err
	}

	backendConfig.Name = backendName
	err = backendConfig.Store(kvs)
	if err != nil {
		log.Warn("Error persisting backend definition")
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	err = kvs.Flush()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

//GetDefinitionList returns a list of backend definitions
func (BackendDef) GetDefinitionList(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	log.Info("backend GetDefinitionList service called")
	backends, err := config.ListBackendConfigs(kvs)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	if backends == nil {
		backends = make([]*config.BackendConfig, 0)
	}

	return backends, nil

}

//GetDefinition retrieves a specific backend definition
func (BackendDef) GetDefinition(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	backendName := resourceIDFromURI(req.RequestURI)

	backendConfig, err := config.ReadBackendConfig(backendName, kvs)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	if backendConfig == nil {
		resp.WriteHeader(http.StatusNotFound)
		return nil, errBackendNotFound
	}

	return backendConfig, err

}

func (BackendDef) DoPost(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	resp.WriteHeader(http.StatusMethodNotAllowed)
	return nil, nil
}
