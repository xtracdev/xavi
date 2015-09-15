package preflocal

import (
	. "github.com/lsegal/gucumber"
	log "github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/xtracdev/xavi/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/xtracdev/xavi/internal/testsupport"
)

//prefLocalImposter is the mountebank imposter configuration used for testing the prefer-local
//load balancer. This configuration is applied to two mountebank servers - one in the same
//docker container as the xavi test endpoint, and the other in a separate container.
const prefLocalImposter = `
	{
  "port": 3001,
  "protocol": "http",
  "stubs": [
    {
      "responses": [
        {
          "is": {
            "statusCode": 200,
            "body": "All work and no play makes Jack a dull boy.\nAll work and no play makes Jack a dull boy.\nAll work and no play makes Jack a dull boy.\nAll work and no play makes Jack a dull boy.\n"
          }
        }
      ],
      "predicates": [
        {
          "equals": {
            "path": "/hello2",
            "method": "GET"
          }
        }
      ]
    }
  ]
}
`

func init() {

	//Definitions used to set up the xavi test server to be spawned
	const (
		prefLocalServer1 = `{"Address":"localhost","Port":3001,"PingURI":"/hello"}`
		prefLocalServer2 = `{"Address":"mbhost","Port":3001,"PingURI":"/hello"}`
		backend          = `{"ServerNames":["local-hello","remote-hello"],"LoadBalancerPolicy":"prefer-local"}`
		route            = `{"URIRoot":"/hello2","Backend":"pref-local-backend","Filters":null,"MsgProps":""}`
		listener         = `{"RouteNames":["pref-local-route"]}`
	)

	//Endpoints associated with the test
	var (
		xaviAgentURL = testsupport.XaviAgentRESTEnpointBaseURI
		testUrl      = testsupport.XaviAcceptanceTestEndpointBaseURL + "/hello2"
		server1Url   = testsupport.CohostedMountebankEndpointBaseURL + "/imposters/3001"
		server2Url   = testsupport.StandaloneMountebackEndpointBaseURL + "/imposters/3001"
	)

	var (
		server1RequestCount int
		server2RequestCount int
		spawnedPID          int
		testFailure         bool
	)

	var doSetup = func() error {

		testPort, err := testsupport.GetPortFromURL(testUrl)
		if err != nil {
			return err
		}

		//
		// Set up XAVI definitions
		//
		log.Info("agent url: ", xaviAgentURL)
		log.Info("server1Url: ", server1Url)

		err = testsupport.PutDefinitionOk("v1/servers/local-hello", prefLocalServer1, xaviAgentURL)
		if err != nil {
			return err
		}

		err = testsupport.PutDefinitionOk("v1/servers/remote-hello", prefLocalServer2, xaviAgentURL)
		if err != nil {
			return err
		}

		err = testsupport.PutDefinitionOk("v1/backends/pref-local-backend", backend, xaviAgentURL)
		if err != nil {
			return err
		}

		err = testsupport.PutDefinitionOk("v1/routes/pref-local-route", route, xaviAgentURL)
		if err != nil {
			return err
		}

		err = testsupport.PutDefinitionOk("v1/listeners/pref-local-listener", listener, xaviAgentURL)
		if err != nil {
			return err
		}

		//
		// Setup the imposters that represent the servers associated with the XAVI backend def
		//
		testsupport.DeleteMountebankImposter(server1Url)
		testsupport.DeleteMountebankImposter(server2Url)
		testsupport.MountebankSetup(testsupport.CohostedMountebankEndpointBaseURL+"/imposters", prefLocalImposter)
		testsupport.MountebankSetup(testsupport.StandaloneMountebackEndpointBaseURL+"/imposters", prefLocalImposter)

		//
		// Spawn the XAVI instance to test with the above configuration
		//
		spawnedPID, err = testsupport.Spawn("pref-local-listener", testPort, xaviAgentURL)
		log.Println("Spawned ", spawnedPID)
		return err
	}

	Given(`^A preflocal route with backend definitions with two servers$`, func() {

		if err := doSetup(); err != nil {
			log.Info("Setup failed: ", err.Error())
			T.Errorf("Error in test setup: %s", err.Error())
			testFailure = true
			return
		}

		//Baseline the request counts at the imposters
		endpointOutput, err := testsupport.GetTestEndpointOutput(server1Url)
		assert.Nil(T, err)
		server1RequestCount = testsupport.CountRequestFrom(endpointOutput)

		endpointOutput, err = testsupport.GetTestEndpointOutput(server2Url)
		assert.Nil(T, err)
		server2RequestCount = testsupport.CountRequestFrom(endpointOutput)

		log.Info("server 1 request count: ", server1RequestCount)
		log.Info("server 2 request count: ", server2RequestCount)

	})

	And(`^The load balancing policy is prefer local$`, func() {
	})

	And(`^I send two requests to the prelocal listener$`, func() {
		if testFailure {
			return
		}
		//Send two requests to the proxies endpoint
		log.Info("send request to endpoint ", testUrl)
		assert.Equal(T, 200, testsupport.GetTestEndpoint(testUrl))
		log.Info("send request")
		assert.Equal(T, 200, testsupport.GetTestEndpoint(testUrl))

	})

	Then(`^Only the local server handles the requests$`, func() {
		if testFailure {
			return
		}

		//Grab the counts and compare them to the baseline - only the endpoint in the
		//container should see requests
		endpointOutput, err := testsupport.GetTestEndpointOutput(server1Url)
		assert.Nil(T, err)
		latestServer1Count := testsupport.CountRequestFrom(endpointOutput)

		endpointOutput, err = testsupport.GetTestEndpointOutput(server2Url)
		assert.Nil(T, err)
		latestServer2Count := testsupport.CountRequestFrom(endpointOutput)

		log.Info("updated server 1 request count", latestServer1Count)
		log.Info("update server 2 request count", latestServer2Count)

		//TODO - better solution is to parse the response from Mountebank as there
		//are two collections the reqeustFrom appears in, which means we see four additional
		//requests in the test which doesn't align nicely with the Gherkin description.
		assert.Equal(T, server1RequestCount+4, latestServer1Count)
		assert.Equal(T, server2RequestCount, latestServer2Count)

	})

	After("@preflocal", func() {
		testPort, err := testsupport.GetPortFromURL(testUrl)
		assert.NotNil(T, err)
		testsupport.KillSpawnedProcess(spawnedPID, testPort, xaviAgentURL)
	})

}
