package acceptancetest

import (
	. "github.com/lsegal/gucumber"
	"github.com/xtracdev/xavi/internal/testsupport"
	"log"
)

//The purpose of this test file is to set the GlobalContext BeforeAll and AfterAll
//functions to do container set up and tear down. Note that these hooks take effect
//even when running tests using tags to filter which tests are run.
func init() {

	var containerCtx map[string]string

	Given(`^Some tests to run$`, func() {
	})

	Then(`^The test containers are started before and stopped after$`, func() {
	})

	GlobalContext.BeforeAll(func() {
		log.Println("starting test containers")
		containerCtx = testsupport.SpawnTestContainers()
	})

	GlobalContext.AfterAll(func() {
		log.Println("stop and remove test containers")
		testsupport.StopAndRemoveContainers(containerCtx)
	})
}
