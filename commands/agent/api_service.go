package agent

import (
	"fmt"
	log "github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/kvstore"
	"net/http"
)

//APICommand defines common functionality that web api enabled configuration
//services must implement
type APICommand interface {
	GetURIRoot() string
	GetDefinition(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error)
	GetDefinitionList(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error)
	PutDefinition(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error)
	DoPost(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error)
}

//APIService exposes APICommand instances as REST services
type APIService struct {
	cmd APICommand
}

//NewAPIService instantiates an API service instance with the given command
func NewAPIService(cmd APICommand) *APIService {
	return &APIService{cmd}
}

func (apiSvc *APIService) serversEndpoint(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	log.Info(fmt.Sprintf("Handling REST request at %s", req.RequestURI))
	list := (req.URL.RequestURI() == apiSvc.cmd.GetURIRoot())

	switch req.Method {
	case "GET":
		if list {
			log.Info("serving list requst")
			return apiSvc.cmd.GetDefinitionList(kvs, resp, req)
		}

		return apiSvc.cmd.GetDefinition(kvs, resp, req)

	case "PUT":
		return apiSvc.cmd.PutDefinition(kvs, resp, req)

	case "POST":
		return apiSvc.cmd.DoPost(kvs, resp, req)

	default:
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return nil, nil
	}
}
