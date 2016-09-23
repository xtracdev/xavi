package roundrobin

import (
	log "github.com/Sirupsen/logrus"
	. "github.com/lsegal/gucumber"
	"github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/internal/testsupport"
)

func init() {

	//Endpoints associated with the test
	var (
		xaviAgentURL = testsupport.XaviAgentRESTEnpointBaseURI
		testUrl      = testsupport.XaviAcceptanceTestEndpointBaseURL + "/hello"
		server1Url   = testsupport.CohostedMountebankEndpointBaseURL + "/imposters/3000"
		server2Url   = testsupport.CohostedMountebankEndpointBaseURL + "/imposters/3100"
	)

	//XAVI definitions for the test scenario
	const (
		hello1Server = `{"Address":"localhost","Port":3000,"PingURI":"/hello"}`
		hello2Server = `{"Address":"localhost","Port":3100,"PingURI":"/hello"}`
		backend      = `{"ServerNames":["hello1","hello2"],"LoadBalancerPolicy":"round-robin"}`
		route        = `{"URIRoot":"/hello","Backends":["demo-backend"],"Plugins":null,"MsgProps":""}`
		listener     = `{"RouteNames":["demo-route"]}`
	)

	var (
		testFailure         bool
		server1RequestCount int
		server2RequestCount int
		spawnedPID          int
	)

	var doSetup = func() error {
		log.Info("set up")
		testPort, err := testsupport.GetPortFromURL(testUrl)
		if err != nil {
			return err
		}

		//
		// XAVI configuration for the test
		//
		err = testsupport.PutDefinitionOk("v1/servers/hello1", hello1Server, xaviAgentURL)
		if err != nil {
			return err
		}

		err = testsupport.PutDefinitionOk("v1/servers/hello2", hello2Server, xaviAgentURL)
		if err != nil {
			return err
		}

		err = testsupport.PutDefinitionOk("v1/backends/demo-backend", backend, xaviAgentURL)
		if err != nil {
			return err
		}

		err = testsupport.PutDefinitionOk("v1/routes/demo-route", route, xaviAgentURL)
		if err != nil {
			return err
		}

		err = testsupport.PutDefinitionOk("v1/listeners/demo-listener", listener, xaviAgentURL)
		if err != nil {
			return err
		}

		//
		// Configuration of the mountebank imposters that represent the servers proxied in
		// the test.
		//
		testsupport.MountebankSetup(testsupport.CohostedMountebankEndpointBaseURL+"/imposters", testsupport.RoundRobin3000Config)
		testsupport.MountebankSetup(testsupport.CohostedMountebankEndpointBaseURL+"/imposters", testsupport.RoundRobin3100Config)

		//
		// Spawn the XAVI process that represents the system under test
		//
		spawnedPID, err = testsupport.Spawn("demo-listener", testPort, xaviAgentURL)
		log.Info("Spawned ", spawnedPID)
		return err
	}

	Given(`^I have a backend definitions with two servers$`, func() {
		if err := doSetup(); err != nil {
			log.Info("Setup failed: ", err.Error())
			T.Errorf("Error in test setup: %s", err.Error())
			testFailure = true
			return
		}

		//Baseline the imposter request counts
		endpointOutput, err := testsupport.GetTestEndpointOutput(server1Url)
		assert.Nil(T, err)
		log.Infof("Counting requests from %s for %s", endpointOutput, server1Url)
		server1RequestCount = testsupport.CountRequestFrom(endpointOutput)

		endpointOutput, err = testsupport.GetTestEndpointOutput(server2Url)
		assert.Nil(T, err)
		server2RequestCount = testsupport.CountRequestFrom(endpointOutput)
		log.Info("server 1 request count: ", server1RequestCount)
		log.Info("server 2 request count: ", server2RequestCount)
	})

	And(`^The load balancing policy is round robin$`, func() {
	})

	And(`^I send two requests to the listener$`, func() {
		if testFailure {
			return
		}
		log.Infof("send request to %s", testUrl)
		assert.Equal(T, 200, testsupport.GetTestEndpoint(testUrl))
		log.Infof("send request to %s", testUrl)
		assert.Equal(T, 200, testsupport.GetTestEndpoint(testUrl))

	})

	Then(`^Each server gets a single request$`, func() {
		if testFailure {
			return
		}

		//Grab the latest request counts for comparison to the baseline counts
		endpointOutput, err := testsupport.GetTestEndpointOutput(server1Url)
		assert.Nil(T, err)
		latestServer1Count := testsupport.CountRequestFrom(endpointOutput)

		endpointOutput, err = testsupport.GetTestEndpointOutput(server2Url)
		assert.Nil(T, err)
		latestServer2Count := testsupport.CountRequestFrom(endpointOutput)

		log.Info("updated server 1 request count: ", latestServer1Count)
		log.Info("update server 2 request count: ", latestServer2Count)

		assert.Equal(T, server1RequestCount+1, latestServer1Count)
		assert.Equal(T, server2RequestCount+1, latestServer2Count)

	})

	After("@basicroundrobin", func() {
		log.Info("After")
		testPort, err := testsupport.GetPortFromURL(testUrl)
		assert.NotNil(T, err)
		testsupport.KillSpawnedProcess(spawnedPID, testPort, xaviAgentURL)
	})

}
