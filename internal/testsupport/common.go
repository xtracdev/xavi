package testsupport

import (
	"fmt"
	. "github.com/lsegal/gucumber"
	log "github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/samalba/dockerclient"
	"github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/internal/testsupport/testcontainer"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	XaviAgentRESTEnpointBaseURI         string
	XaviAcceptanceTestEndpointBaseURL   string
	StandaloneMountebackEndpointBaseURL string
	CohostedMountebankEndpointBaseURL   string
)

func init() {

	/*
				Environment samples

		export XAVI_AT_AGENT_ENDPOINT=http://localhost:8080
		export XAVI_AT_TEST_ENDPOINT=http://localhost:9000
		export XAVI_AT_STANDALONE_MB=http://localhost:3636
		export XAVI_AT_COHOSTED_MB=http://localhost:3535

		export XAVI_AT_AGENT_ENDPOINT=http://vc2coma2046699n:8080
		export XAVI_AT_TEST_ENDPOINT=http://vc2coma2046699n:9000
		export XAVI_AT_STANDALONE_MB=http://vc2coma2046699n:2626
		export XAVI_AT_COHOSTED_MB=http://vc2coma2046699n:2525

	*/

	XaviAgentRESTEnpointBaseURI = os.Getenv("XAVI_AT_AGENT_ENDPOINT")
	if XaviAgentRESTEnpointBaseURI == "" {
		XaviAgentRESTEnpointBaseURI = "http://localhost:8080"
	}

	XaviAcceptanceTestEndpointBaseURL = os.Getenv("XAVI_AT_TEST_ENDPOINT")
	if XaviAcceptanceTestEndpointBaseURL == "" {
		XaviAcceptanceTestEndpointBaseURL = "http://localhost:9000"
	}

	StandaloneMountebackEndpointBaseURL = os.Getenv("XAVI_AT_STANDALONE_MB")
	if StandaloneMountebackEndpointBaseURL == "" {
		StandaloneMountebackEndpointBaseURL = "http://localhost:3636"
	}

	CohostedMountebankEndpointBaseURL = os.Getenv("XAVI_AT_COHOSTED_MB")
	if CohostedMountebankEndpointBaseURL == "" {
		CohostedMountebankEndpointBaseURL = "http://localhost:3535"
	}

}

//GetTestEndpoint returns the HTTP status code obtained from doing
//an HTTP get on the supplied endpoint
func GetTestEndpoint(endpoint string) int {
	return GetTestEndpointWithHeader(endpoint, "", "")
}

func GetTestEndpointWithHeader(endpoint string, header string, value string) int {

	client := &http.Client{}
	req, err := http.NewRequest("GET", endpoint, nil)
	assert.Nil(T, err)

	if header != "" {
		req.Header.Add(header, value)
	}

	resp, err := client.Do(req)
	assert.Nil(T, err)

	log.Info("GetTestEndpoint - endpoint returned ", resp)
	//The nil test is probably not needed, but was added after seeing the first get to a test
	//endpoint that was spawned in a docker container produce a nil response. This was eliminated
	//by sleeping one second after spawning the xavi test endpoint in the docker container, and
	//likely represented a race condition between the get on the spawned endpoint vs the port forwarding
	//path (host os -> vagrant guest os -> docker container) establishment.
	if resp != nil {
		return resp.StatusCode
	} else {
		log.Info("Null response in SentTestRequest")
		return http.StatusInternalServerError
	}
}

//GetTestEndpointOutput returns the response obtained via an HTTP Get on endpoint
func GetTestEndpointOutput(endpoint string) (string, error) {
	resp, err := http.Get(endpoint)
	if resp == nil {
		return "", fmt.Errorf("Empty response returned from endpoint")
	}

	if err != nil {
		return "", err
	}
	respPayload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil
	}
	resp.Body.Close()
	return string(respPayload), nil
}

//CountRequestFrom returns the number of occurances of "requestFrom:" is the
//given string. This is useful when looking at imposter stats obtained from a mountebank
//server.
func CountRequestFrom(s string) int {
	return strings.Count(s, "\"requestFrom\":")
}

//PutDefinition is used to put an entity to an HTTP endpoint. This is used to put
//XAVI setup definitions to an endpoint.
func PutDefinition(serviceURI string, payload string, agentURL string) (int, error) {
	testURL := fmt.Sprintf("%s/%s", agentURL, serviceURI)
	log.Info("testURL is ", testURL)
	request, err := http.NewRequest("PUT", testURL, strings.NewReader(payload))
	if err != nil {
		return -1, err
	}

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		return -1, err
	}

	log.Info(fmt.Sprintf("put def response: %v", response))
	return response.StatusCode, nil
}

//PutDefinitionOk calls PutDefinition, returning an error if PutDefinition returns
//and error or a status code other than StatusOk
func PutDefinitionOk(serviceURI string, payload string, agentURL string) error {
	code, err := PutDefinition(serviceURI, payload, agentURL)
	if err != nil {
		return err
	}

	if code != http.StatusOK {
		return fmt.Errorf("Non-OK status returned: %d", code)
	}

	return nil
}

//Spawn used the XAVI spawn-listener endpoint to start an instance of XAVI for testing. The
//spawned process uses the definitions associated with the named listener.
func Spawn(name string, port int, agentURL string) (int, error) {

	spawnUrl := fmt.Sprintf("%s/v1/spawn-listener/", agentURL)

	//Note we use the 0.0.0.0 address in the docker container to ensure the port can be forwarded to
	//from the docker container host.
	payload := `{"ListenerName":"` + name + `", "Address":"` + "0.0.0.0:" + fmt.Sprintf("%d", port) + `"}`

	resp, err := http.Post(spawnUrl, "application/json", strings.NewReader(payload))
	if err != nil {
		return -1, err
	}

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}
	resp.Body.Close()
	log.Info("Spawn response: ", string(out))
	pid, err := parseSpawnedPID(string(out))
	if err != nil {
		return -1, err
	}

	//May need a slight delay to let the vagrant/docker register the spawned tcp listener. Without the delay the
	//first test against a spawned endpoint fails, but all subsequent test runs pass. The failure
	//is caused by a nil access panic reading the response code.
	time.Sleep(1 * time.Second)
	return pid, err

}

//KillSpawnedProcess uses the spawn-killer endpoint in the xavi agent to kill
//a spawned process. This is useful in cleaning up processes that were spawned
//to execute a test case.
func KillSpawnedProcess(pid int, port int, agentURL string) error {
	spawnUrl := fmt.Sprintf("%s/v1/spawn-killer/%d", agentURL, pid)

	resp, err := http.Post(spawnUrl, "", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpect status code returned after killing spawned process: %d", resp.StatusCode)
	}

	return nil
}

//MountebankSetup is used to post mountebank setup configuration to a mountebank endpoint.
func MountebankSetup(endpoint string, config string) error {
	request, err := http.NewRequest("POST", endpoint, strings.NewReader(config))
	if err != nil {
		log.Warn("Error creating request for mountebank setup: ", err.Error())
		return err
	}

	request.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Warn("Error returned by client.Do in mountebank setup: ", err.Error())
		return err
	}

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warn("Error reading response body in mountebank setup: ", err.Error())
		return err
	}
	resp.Body.Close()
	log.Info("Mountebank response: ", string(out))

	return nil

}

//parseSpawnedPID parses the pid from the spawn response
func parseSpawnedPID(response string) (int, error) {
	parts := strings.Split(response, " ")
	if len(parts) != 3 {
		return -1, fmt.Errorf("Expected content to look like 'started process <pid>'")
	}

	pidPart := parts[2]

	//Sometimes the response is contained in double quotes, which means the last part
	//ends with a trailing double quote
	if strings.Contains(pidPart, "\"") {
		pidPart = strings.Split(pidPart, "\"")[0]
	}

	return strconv.Atoi(pidPart)
}

func GetPortFromURL(urlString string) (int, error) {
	theUrl, err := url.Parse(urlString)
	if err != nil {
		return -1, err
	}

	parts := strings.Split(theUrl.Host, ":")
	switch len(parts) {
	case 1:
		return 80, nil
	case 2:
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return -1, err
		}
		return port, nil
	default:
		return -1, fmt.Errorf("Unable to glean port from URL")
	}

}

//DeleteMountebankImposter is used to delete a Mountebank imposter via a delete to an imposter resource
func DeleteMountebankImposter(url string) {
	req, err := http.NewRequest("DELETE", url, nil)
	assert.Nil(T, err)

	client := &http.Client{}
	client.Do(req)
}

const (
	XaviTestContainer       = "xavi"
	MountebankTestContainer = "mountebank"
)

func createMountebankTestContainerContext() *testcontainer.ContainerContext {
	containerCtx := testcontainer.ContainerContext{
		ImageName: "mb-server-alpine",
	}

	containerCtx.Labels = make(map[string]string)
	containerCtx.Labels["xt-container-type"] = "atest-mb"

	containerCtx.PortContext = make(map[string]string)
	containerCtx.PortContext["2525/tcp"] = "2626"

	return &containerCtx
}

func createXaviTestContainerContext(linkedContainerName string) *testcontainer.ContainerContext {
	containerCtx := testcontainer.ContainerContext{
		ImageName: "xavi-test-alpine-base",
	}

	containerCtx.Labels = make(map[string]string)
	containerCtx.Labels["xt-container-type"] = "atest-xavi"

	containerCtx.PortContext = make(map[string]string)
	containerCtx.PortContext["2525/tcp"] = "2525"
	containerCtx.PortContext["8080/tcp"] = "8080"
	containerCtx.PortContext["9000/tcp"] = "9000"

	containerCtx.Links = append(containerCtx.Links, fmt.Sprintf("%s:mbhost", linkedContainerName[1:]))
	log.Println(fmt.Sprintf("%v", containerCtx.Links))

	return &containerCtx
}

//SpawnTestContainers checks the status of the required test containers that must be running
//in able for the acceptance tests to execute successfully. If the containers are not running then
//they are started. The map that is returned containers container name to container id mapping.
func SpawnTestContainers() map[string]string {
	var bootedOneOrMoreContainers bool

	containerMapping := make(map[string]string)

	//Grab the environment
	dockerHost, dockerCertPath := testcontainer.ReadDockerEnv()

	// Init the client
	log.Println("Create docker client")
	docker, _ := dockerclient.NewDockerClient(dockerHost, testcontainer.BuildDockerTLSConfig(dockerCertPath))

	// Is the mountebank container already running?
	log.Println("Check to see if mountebank test container is already started")
	info := testcontainer.GetAcceptanceTestContainerInfo(docker, "atest-mb")
	if info != nil {
		log.Println("Mountebank container found - state is: ", info.State.StateString())
		containerMapping[MountebankTestContainer] = info.Id
	} else {
		log.Println("Mountebank container not running - create container context")
		bootedOneOrMoreContainers = true
		containerCtx := createMountebankTestContainerContext()

		//Create and start the container.
		log.Println("Create and start the container")
		mountebankContainerId := testcontainer.CreateAndStartContainer(docker, containerCtx)
		containerMapping[MountebankTestContainer] = mountebankContainerId

		//Need to get test container info after start as name needed to link second container to it.
		info = testcontainer.GetAcceptanceTestContainerInfo(docker, "atest-mb")
	}

	var mbContainerName = info.Name

	// Is the xavi test container already running?
	log.Println("Check to see if xavi test container is already started")
	info = testcontainer.GetAcceptanceTestContainerInfo(docker, "atest-xavi")
	if info != nil {
		log.Println("Xavi test container found - state is: ", info.State.StateString())
		log.Println("Xavi container links - ", info.HostConfig.Links)
		containerMapping[XaviTestContainer] = info.Id
	} else {
		log.Println("Xavi test container not running - create container context")
		bootedOneOrMoreContainers = true
		xaviContainerCtx := createXaviTestContainerContext(mbContainerName)

		//Create and start the container.
		log.Println("Create and start the container")
		xaviTestContainerId := testcontainer.CreateAndStartContainer(docker, xaviContainerCtx)
		containerMapping[XaviTestContainer] = xaviTestContainerId
	}

	// Pause to let the containers boot
	if bootedOneOrMoreContainers {
		//If the test hits the xavi container right away the test set up fails, as the container state
		//can be running before xavi is ready to accept traffic.
		//TODO - add a 'ping' call in the common test library to get a 200 response on a config API list
		//service, ignoring errors - once 200 returns then the service is available and tests can proceed.
		time.Sleep(1 * time.Second)
	}

	return containerMapping
}

//StopAndRemoveContainers stops and removes the containers in the given map
func StopAndRemoveContainers(containers map[string]string) {
	if len(containers) == 0 {
		return
	}

	if os.Getenv("XT_CLEANUP_CONTAINERS") == "false" {
		log.Info("Skipping container stop and remove - XT_CLEANUP_CONTAINERS is false")
		return
	}

	//Grab the environment
	dockerHost, dockerCertPath := testcontainer.ReadDockerEnv()

	// Init the client
	docker, _ := dockerclient.NewDockerClient(dockerHost, testcontainer.BuildDockerTLSConfig(dockerCertPath))

	log.Info("stop test containers")

	//Note that since xavi is linked to mountebank we stop/remove xavi first
	log.Info("stop xavi test container")
	docker.StopContainer(containers[XaviTestContainer], 5)
	log.Info("stop mountebank test container")
	docker.StopContainer(containers[MountebankTestContainer], 5)

	log.Info("remove the test containers")
	docker.RemoveContainer(containers[XaviTestContainer], false, false)
	docker.RemoveContainer(containers[MountebankTestContainer], false, false)

}
