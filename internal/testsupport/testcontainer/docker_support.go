package testcontainer

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/samalba/dockerclient"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

//ReadDockerEnv reads docker environment variables needed by the docker APIs
func ReadDockerEnv() (string, string) {

	dockerHome := os.Getenv("DOCKER_HOST")
	if dockerHome == "" {
		log.Fatal("DOCKER_HOST environment variable not set.")
	}

	dockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	if dockerCertPath == "" {
		log.Fatal("DOCKER_CERT_PATH environment variable not set.")
	}

	return dockerHome, dockerCertPath
}

//BuildDockerTLSConfig builds the docker TLS configuration needed by the API
func BuildDockerTLSConfig(dockerCertPath string) *tls.Config {

	caFile := fmt.Sprintf("%s/ca.pem", dockerCertPath)
	certFile := fmt.Sprintf("%s/cert.pem", dockerCertPath)
	keyFile := fmt.Sprintf("%s/key.pem", dockerCertPath)

	tlsConfig := &tls.Config{}

	cert, _ := tls.LoadX509KeyPair(certFile, keyFile)
	pemCerts, _ := ioutil.ReadFile(caFile)

	tlsConfig.RootCAs = x509.NewCertPool()
	tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	tlsConfig.Certificates = []tls.Certificate{cert}

	tlsConfig.RootCAs.AppendCertsFromPEM(pemCerts)

	return tlsConfig
}

//GetAcceptanceTestContainerInfo returns information about the specified container type
func GetAcceptanceTestContainerInfo(docker *dockerclient.DockerClient, containerType string) *dockerclient.ContainerInfo {

	// Get only running containers
	containers, err := docker.ListContainers(false, false, "")
	if err != nil {
		log.Fatal(err)
	}

	//Loop through them until we find a match
	for _, c := range containers {
		xtContainerType, ok := c.Labels["xt-container-type"]
		if ok && xtContainerType == containerType {
			//Grab the information for the container
			info, err := docker.InspectContainer(c.Id)
			if err != nil {
				log.Fatal(err)
			}

			return info
		}
	}

	return nil
}

//ContainerContext defines context relevant to creating containers via this package
type ContainerContext struct {
	ImageName string
	Labels    map[string]string
	//PortContext has a container port/proto key and a host port value,
	//with a convention that the container port/proto is an exposed port
	//from the container, and a host port it is mapped to is specified
	//in the map. We further restrict things by assuming a single host
	//mapping for an exposed port.
	PortContext map[string]string
	Links       []string
}

func createContainer(docker *dockerclient.DockerClient, ctx *ContainerContext) (string, error) {
	//Make a collection of exposed ports
	var exposedPorts map[string]struct{}
	exposedPorts = make(map[string]struct{})
	for k := range ctx.PortContext {
		exposedPorts[k] = struct{}{}
	}

	//Build the Docker container config from the configuration provided by the caller
	containerConfig := dockerclient.ContainerConfig{
		Image:        ctx.ImageName,
		Labels:       ctx.Labels,
		ExposedPorts: exposedPorts,
	}

	//Create the container
	return docker.CreateContainer(&containerConfig, "")

}

func startContainer(docker *dockerclient.DockerClient, containID string, ctx *ContainerContext) error {
	//Build the port bindings needed when running the container
	dockerHostConfig := new(dockerclient.HostConfig)
	dockerHostConfig.PortBindings = make(map[string][]dockerclient.PortBinding)
	for k, v := range ctx.PortContext {
		pb := dockerclient.PortBinding{HostPort: v}
		portBindings := []dockerclient.PortBinding{pb}
		dockerHostConfig.PortBindings[k] = portBindings
	}

	dockerHostConfig.Links = ctx.Links

	//Start the container
	return docker.StartContainer(containID, dockerHostConfig)
}

//CreateAndStartContainer creates abd starts the container given the context specifying the desired container properties
func CreateAndStartContainer(docker *dockerclient.DockerClient, ctx *ContainerContext) string {

	//Make sure the required image is present
	imagePresent := requiredImageAvailable(docker, ctx.ImageName)
	if !imagePresent {
		log.Fatal("Cannot run test as required image (", ctx.ImageName, ") is not present.")
	}

	//Create the container
	containerID, err := createContainer(docker, ctx)
	if err != nil {
		log.Fatal(err)
	}

	//Start the container
	err = startContainer(docker, containerID, ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("...container started...")

	return containerID
}

func requiredImageAvailable(docker *dockerclient.DockerClient, name string) bool {
	images, err := docker.ListImages(true)
	if err != nil {
		log.Fatal(err)
	}

	for _, i := range images {
		for _, t := range i.RepoTags {
			if strings.Index(t, name) == 0 {
				return true
			}
		}
	}

	return false
}
