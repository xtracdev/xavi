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

const (
	listenersURI = "/v1/listeners/"
)

//Errors
var (
	errListenerNotFound        = errors.New("Listener definition not found")
	errListenerResourceMissing = errors.New("Listener resource not present in url - expected /v1/listeners/listener-resource")
)

//ListenerDefCmd is the interface instance used to expose as an API endpoint
var ListenerDefCmd ListenerDef

//ListenerDef is used to hang the ApiCommand functions needed for exposing listener def capabilities
//via a REST API
type ListenerDef struct{}

//GetURIRoot returns the URI root used to serve listener def API calls
func (ListenerDef) GetURIRoot() string {
	return listenersURI
}

//PutDefinition creates or updates a listener definition
func (ListenerDef) PutDefinition(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	body, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	log.Info(fmt.Sprintf("Put request with payload %s", string(body)))
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, nil
	}

	listenerName := resourceIDFromURI(req.RequestURI)
	log.Info(listenerName)
	if listenerName == "" {
		resp.WriteHeader(http.StatusNotFound)
		return nil, errServiceResourceMissing
	}

	listenerConfig := &config.ListenerConfig{
		HealthEndpoint: true,
	}
	err = json.Unmarshal(body, listenerConfig)
	if err != nil {
		log.Warn("Error unmarshaling request body")
		resp.WriteHeader(http.StatusBadRequest)
		return nil, err
	}

	listenerConfig.Name = listenerName
	err = listenerConfig.Store(kvs)
	if err != nil {
		log.Warn("Error persisting listener definition")
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	err = kvs.Flush()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

//GetDefinitionList returns a list of listener definitions
func (ListenerDef) GetDefinitionList(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	log.Info("listener GetDefinitionList service called")
	listeners, err := config.ListListenerConfigs(kvs)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	if listeners == nil {
		listeners = make([]*config.ListenerConfig, 0)
	}

	return listeners, nil

}

//GetDefinition retrieves a specific listener definition
func (ListenerDef) GetDefinition(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	listenerName := resourceIDFromURI(req.RequestURI)

	listenerConfig, err := config.ReadListenerConfig(listenerName, kvs)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	if listenerConfig == nil {
		resp.WriteHeader(http.StatusNotFound)
		return nil, errListenerNotFound
	}

	return listenerConfig, err

}

//DoPost handles post requests, which are not supported for ListenerDef.
func (ListenerDef) DoPost(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	resp.WriteHeader(http.StatusMethodNotAllowed)
	return nil, nil
}
