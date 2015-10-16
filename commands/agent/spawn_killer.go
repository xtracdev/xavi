package agent

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/kvstore"
	"net/http"
	"os/exec"
	"strconv"
)

const (
	spawnKillURI = "/v1/spawn-killer/"
)

//Exported SpawnKillerDef for external reference
var SpawnKillerDefCmd SpawnKillerDef

//SpawnKillerDef is used to hang the ApiCommand functions needed for killing spawned processes
//via a REST API
type SpawnKillerDef struct{}

//GetURIRoot returns the URI root used to kill spawn xavi listeners
func (SpawnKillerDef) GetURIRoot() string {
	return spawnKillURI
}

//PutDefinition is not implemented for spawn killer
func (SpawnKillerDef) PutDefinition(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	resp.WriteHeader(http.StatusMethodNotAllowed)
	return nil, nil
}

//GetDefinition is not implemented for spawn killer
func (SpawnKillerDef) GetDefinition(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	resp.WriteHeader(http.StatusMethodNotAllowed)
	return nil, nil
}

//GetDefinitionList is not implemented for spawn killer
func (SpawnKillerDef) GetDefinitionList(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	resp.WriteHeader(http.StatusMethodNotAllowed)
	return nil, nil
}

func (SpawnKillerDef) DoPost(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	pidToKill := resourceIDFromURI(req.RequestURI)
	log.Info("request kill of pid ", pidToKill)

	pid, err := strconv.Atoi(pidToKill)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return nil, err
	}

	if !isSpawnedPid(pid) {
		myError := fmt.Errorf("Asked to kill pid %d, which we did not spawn", pid)
		log.Warn(myError.Error())
		resp.WriteHeader(http.StatusBadRequest)
		return nil, myError
	}

	log.Info("Killing ", pid)
	cmd := exec.Command("kill", pidToKill)

	err = cmd.Run()
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	removePid(pid)
	log.Info("Stone cold killa finished")

	return nil, nil
}
