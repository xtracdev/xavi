package msgpropmatch

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

	const (
		server1 = `{"Name":"hello1","Address":"localhost","Port":3000,"PingURI":"/hello"}`
		server2 = `{"Name":"hello2","Address":"localhost","Port":3100,"PingURI":"/hello"}`

		backend1 = `{"Name":"demo-backend-1","ServerNames":["hello1"],"LoadBalancerPolicy":""}`
		backend2 = `{"Name":"demo-backend-2","ServerNames":["hello2"],"LoadBalancerPolicy":""}`

		route1 = `{"Name":"demo-route-1","URIRoot":"/hello","Backends":["demo-backend-1"],"Plugins":null,"MsgProps":"SOAPAction=foo"}`
		route2 = `{"Name":"demo-route-2","URIRoot":"/hello","Backends":["demo-backend-2"],"Plugins":null,"MsgProps":"SOAPAction=bar"}`

		listener = `{"Name":"demo-listener","RouteNames":["demo-route-1","demo-route-2"]}`
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
		err = testsupport.PutDefinitionOk("v1/servers/hello1", server1, xaviAgentURL)
		if err != nil {
			return err
		}

		err = testsupport.PutDefinitionOk("v1/servers/hello2", server2, xaviAgentURL)
		if err != nil {
			return err
		}

		err = testsupport.PutDefinitionOk("v1/backends/demo-backend-1", backend1, xaviAgentURL)
		if err != nil {
			return err
		}

		err = testsupport.PutDefinitionOk("v1/backends/demo-backend-2", backend2, xaviAgentURL)
		if err != nil {
			return err
		}

		err = testsupport.PutDefinitionOk("v1/routes/demo-route-1", route1, xaviAgentURL)
		if err != nil {
			return err
		}

		err = testsupport.PutDefinitionOk("v1/routes/demo-route-2", route2, xaviAgentURL)
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

	Given(`^Routes with msgprop expressions$`, func() {
		if err := doSetup(); err != nil {
			log.Info("Setup failed: ", err.Error())
			T.Errorf("Error in test setup: %s", err.Error())
			testFailure = true
			return
		}

		//Baseline the imposter request counts
		endpointOutput, err := testsupport.GetTestEndpointOutput(server1Url)
		assert.Nil(T, err)
		server1RequestCount = testsupport.CountRequestFrom(endpointOutput)

		endpointOutput, err = testsupport.GetTestEndpointOutput(server2Url)
		assert.Nil(T, err)
		server2RequestCount = testsupport.CountRequestFrom(endpointOutput)
		log.Info("server 1 request count: ", server1RequestCount)
		log.Info("server 2 request count: ", server2RequestCount)
	})

	And(`^The routes have a common uri$`, func() {

	})

	Then(`^Requests are dispatched based on msgprop matching$`, func() {
		if testFailure {
			return
		}

		log.Info("get ", testUrl)
		assert.Equal(T, 200, testsupport.GetTestEndpointWithHeader(testUrl, "SOAPAction", "foo"))
		log.Info("get ", testUrl)
		assert.Equal(T, 200, testsupport.GetTestEndpointWithHeader(testUrl, "SOAPAction", "bar"))

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

	After("@msgpropmatch", func() {
		testPort, err := testsupport.GetPortFromURL(testUrl)
		assert.NotNil(T, err)
		testsupport.KillSpawnedProcess(spawnedPID, testPort, xaviAgentURL)
	})

}
