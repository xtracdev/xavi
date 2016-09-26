## Developing Xavi

### Dependency management

Compile dependencies except for golang.org/x/net/context are managed via [Godep](https://github.com/tools/godep) using golang 1.5 vendoring support. Note the
role of godep is for maintaining the state of the vendor directory; go commands do not need to be
prefaced with go dep. Get the context dependency via `go get golang.org/x/net/context`

Refer to the godep documentation for how to manage dependencies.

When working with this repository, be sure to export GO15VENDOREXPERIMENT=1

Note when applying commands to all subdirectories, you may wish to exclude the vendor subdirectory. For example,
when running tests, trying to run tests in the vendor subdirectory will fail because the test dependencies for
vendored libraries are not included.

To exlude the vendor directory when running tests, do this:

<pre>
go test $(go list ./... | grep -v /vendor/)
</pre>

If modifying or adding a dependency, the path of least resistence seems to be unset GO15VENDOREXPERIMENT, then restore
the environment using godep. Remove Godeps and vendor, do your work, then set GO15VENDOREXPERIMENT and save your 
dependencies. Refer to the godep documentation for details.


### Codeship Build Setup

Setup Commands

<pre>
cd $HOME
wget https://storage.googleapis.com/golang/go1.5.1.linux-amd64.tar.gz
tar xvzf go1.5.1.linux-amd64.tar.gz
export GOROOT=$HOME/go
export PATH=$GOROOT/bin:$PATH
export GO15VENDOREXPERIMENT=1
go version
cd $GOPATH/src/github.com/xtracdev/xavi
</pre>

Test Commands

<pre>
go test $(go list ./... | grep -v /vendor/)
</pre>


### KVStore

XAVI can work with consul as the configuration store, or use a hashmap-based KVStore that can dump and load its
contents to/from a file.

Use the XAVI_KVSTORE_URL environment variable to set the KVStore URL. For consul, the value of the environment
variable should be `consul://host:port` and for a the hashmap/file store specify a file URL for the hashmap backing
file. Note that file URLs are full paths to files, e.g.

<pre>
export XAVI_KVSTORE_URL=file:////some/path/democfg.xavi
</pre>



### Logging

[Logrus](https://github.com/Sirupsen/logrus) is currently the logging framework. The log level can be set
via the XAVI_LOGGING_LEVEL environment variable (valid values are debug, info, warn, fatal, error, panic).



### Cross-compiling

For details on cross compiling see https://gist.github.com/d-smith/9d7ca1baa72135dfe7b0

TL;DR

<pre>
GOOS=linux GOARCH=386 CGO_ENABLED=0 go build
</pre>



### Go Code Coverage

		go test -coverprofile=coverage.out; go tool cover -html=coverage.out
		
#### Go Code Coverage with Gocov

Gocov will accumulate coverage recursively, unlike the go test tool which produces coverage for a single
package.

        gocov test $(go list ./... | grep -v /vendor/) |gocov-html > coverage.html

### Port Usage - Mac Os X  

	lsof -n -i4TCP

### Acceptance Testing Setup with Docker for Mac

Assuming you are working on a mac, install Docker for Mac to use the docker based
gucumber tests. Docker for Mac works pretty good, definitely friendlier than
docker machine on the Mac now that it's out of beta.

Once docker machine is configured correctly, you can run the acceptance tests 
using gucumber. Note that this version vendors an older version of gucumber, so
do a `go install` in the appropriate vendor directory to put the right
version of the command in your path.

#### Mountebank

Note the above port forwarding - it maps the host os perspective to the guest os
perspective, with docker port forwarding managing the mapping of the guest os to the
container. 

Also note that there is no need to start the containers using docker - the acceptance tests will
check to see if the container is running. If the container is not running, the acceptance test will
create and start it using the docker API. The commands below are provided as documentation of what the
acceptance test code does to run the containers.

Build the Mountebank server image in the docker/mountebank-alpine directory:

<pre>
docker build -t "mb-server-alpine" .
</pre>

You can run it like this:

</pre>
docker run -d -p 2626:2525 --name mountebank --label 'xt-container-type=atest-mb' mb-server-alpine
</pre>

If you are behind a proxy pass the proxy info via build args, e.g. `docker build --build-arg HTTP_PROXY=something --build-arg HTTPS_PROXY=etc.`

#### XAVI Docker set up

Note that you must first cross-compile xavi for linux and copy it into docker/xavi-alpine 
before building the image, e.g. `GOOS=linux GOARCH=386 CGO_ENABLED=0 go build`

Build the image in the docker/xavi-alpine directory

<pre>
docker build -t "xavi-test-alpine-base" .
docker run -d -p 8080:8080 -p 9000:9000 -p 2525:2525 --name xavi-docker --label 'xt-container-type=atest-xavi' --link mountebank:mbhost xavi-test-alpine-base
</pre>


By default the containers are stopped at the end of each test run, which is best for ensure clean test environments in
a known state, as opposed to containers in indeterminate states, for example if test failures leaves ports tied up,
test configuration that can't be cleaned up, etc.

For the purpose of quickly running tests that are meant to test the introduction of breaking changes, the containers
can be left running if gucumber is executued with the environment variable XT_CLEANUP_CONTAINERS=false, e.g.

<pre>
env XT_CLEANUP_CONTAINERS=false gucumber
</pre>
