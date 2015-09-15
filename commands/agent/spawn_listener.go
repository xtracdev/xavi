package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/kvstore"
	"io/ioutil"
	"net/http"
	"os/exec"
	"sync"
)

const (
	spawnURI = "/v1/spawn-listener/"
)

//We keep track of the pids we spawn as we only kill what we spawn.
var (
	myPids   map[int]int
	mapMutex sync.Mutex
)

func init() {
	myPids = make(map[int]int)
}

func addPid(pid int) {
	mapMutex.Lock()
	myPids[pid] = pid
	mapMutex.Unlock()
}

func removePid(pid int) {
	mapMutex.Lock()
	delete(myPids, pid)
	mapMutex.Unlock()
}

func isSpawnedPid(pid int) bool {
	mapMutex.Lock()
	_, ok := myPids[pid]
	mapMutex.Unlock()
	return ok
}

//Exported SpawnListenerDefCmd for external reference
var SpawnListenerDefCmd SpawnListenerDef

//SpawnListenerDef is used to hang the ApiCommand functions needed for spawning xavi instances
type SpawnListenerDef struct{}

//GetURIRoot returns the URI root used to spawn listeners
func (SpawnListenerDef) GetURIRoot() string {
	return spawnURI
}

//Listeners are spawned via the specification of a listener name and address
type spawn struct {
	ListenerName string
	Address      string
}

//PutDefinition not provided for spawning
func (SpawnListenerDef) PutDefinition(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	resp.WriteHeader(http.StatusMethodNotAllowed)
	return nil, nil
}

//GetDefinition is not provided for spawning
func (SpawnListenerDef) GetDefinition(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	resp.WriteHeader(http.StatusMethodNotAllowed)
	return nil, nil
}

//GetDefinitionList is not provided for spawning
func (SpawnListenerDef) GetDefinitionList(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	resp.WriteHeader(http.StatusMethodNotAllowed)
	return nil, nil
}

//DoPost spawns the xavi listener as specified by the payload
func (SpawnListenerDef) DoPost(kvs kvstore.KVStore, resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	body, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	log.Info(fmt.Sprintf("Put request with payload %s", string(body)))

	spawnSpec := new(spawn)
	err = json.Unmarshal(body, spawnSpec)
	if err != nil {
		log.Warn("Error unmarshaling request body")
		resp.WriteHeader(http.StatusBadRequest)
		return nil, err
	}

	log.Info("Spawning a process - form command")
	var pout, perror bytes.Buffer
	cmd := exec.Command("xavi", "listen", "-ln", spawnSpec.ListenerName, "-address", spawnSpec.Address)
	cmd.Stderr = &perror
	cmd.Stdout = &pout

	log.Info("run command")
	err = cmd.Start()
	if err != nil {
		log.Warn("error running command: ", err.Error())
		resp.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	pid := cmd.Process.Pid
	log.Info("started process - pid is: ", pid)
	addPid(pid)

	return fmt.Sprintf("started process %d", pid), nil
}
