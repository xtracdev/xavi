## Xavi Quick Start

This guide shows a simple example of proxying an endpoint.

### Demo Service With Mountebank

We can use [Mountebank](http://www.mbtest.org/) as a service endpoint for trying out XAVI.

Consider the following mountebank imposter definition (see democonfig/imposter.json):

<pre>
{
  "port": 4545,
  "protocol": "http",
  "stubs": [
    {
      "responses": [
        {
          "is": {
            "statusCode": 400,
            "body": "All work and no play makes Jack a dull boy.\nAll work and no play makes Jack a dull boy.\nAll work and no play makes Jack a dull boy.\nAll work and no play makes Jack a dull boy.\n"
          }
        }
      ],
      "predicates": [
        {
          "equals": {
            "path": "/hello",
            "method": "GET"
          }
        }
      ]
    }
  ]
}
</pre>

We can provide a mock /hello service by defining it using Mountebank:

<pre>
curl -i -X POST -H 'Content-Type: application/json' -d@democonfig/imposter.json http://127.0.0.1:2525/imposters
</pre>

The service endpoint can be called via curl:

<pre>
curl localhost:4545/hello
All work and no play makes Jack a dull boy.
All work and no play makes Jack a dull boy.
All work and no play makes Jack a dull boy.
All work and no play makes Jack a dull boy.
</pre>

We can then set up a simple proxy example like this:

<pre>
export XAVI_KVSTORE_URL=file:////some/path/mbdemo.xavi
curl -i -X POST -H 'Content-Type: application/json' -d@democonfig/imposter.json http://127.0.0.1:2525/imposters
./xavi add-server -address localhost -port 4545 -name hello1 -ping-uri /hello
./xavi add-backend -name demo-backend -servers hello1
./xavi add-route -name demo-route -backend demo-backend -base-uri /hello
./xavi add-listener -name demo-listener -routes demo-route
./xavi listen -ln demo-listener -address 0.0.0.0:8080
</pre>

After the listener is started, the proxy endpoint can be used, e.g. `curl localhost:8080/hello`

For a two server round-robin proxy config demo, try this:

<pre>
export XAVI_KVSTORE_URL=file:////some/path/mbdemo.xavi
curl -i -X POST -H 'Content-Type: application/json' -d@democonfig/hello3000.json http://127.0.0.1:2525/imposters
curl -i -X POST -H 'Content-Type: application/json' -d@democonfig/hello3100.json http://127.0.0.1:2525/imposters
./xavi add-server -address localhost -port 3000 -name hello1 -ping-uri /hello -health-check http-get
./xavi add-server -address localhost -port 3100 -name hello2 -ping-uri /hello -health-check http-get
./xavi add-backend -name demo-backend -servers hello1,hello2 -load-balancer-policy round-robin
./xavi add-route -name demo-route -backend demo-backend -base-uri /hello
./xavi add-listener -name demo-listener -routes demo-route
./xavi listen -ln demo-listener -address 0.0.0.0:8080
</pre>


Curling the endpoint as in the single server instance, the mountebank log shows the requests
being distributed to the two endpoints.

<pre>
info: [http:3000] ::ffff:127.0.0.1:50568 => GET /hello
info: [http:3100] ::ffff:127.0.0.1:50571 => GET /hello
info: [http:3000] ::ffff:127.0.0.1:50573 => GET /hello
info: [http:3100] ::ffff:127.0.0.1:50575 => GET /hello
</pre>

### REST Facade Example

The [Xavi Sample](https://github.com/xtracdev/xavi-sample) project shows how to use the Xavi toolkit to create a RESTful
facade on top of a SOAP service.

The Xavi sample also shows how to integrate with an HTTPs backend, and how to deal with timeouts and
cancellations.