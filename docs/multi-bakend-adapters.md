# Multi-Backend Adapters

## Overview

In some cases, an API facade implemented using Xavi might aggregate data obtained via HTTP calls to 
multiple backends. While HTTP calls could be made from a plugin using the HTTP client, there are 
a number of problems associated with direct calls, such as not leveraging load balancers, requiring knowledge of
server endpoints in the plugin, not being able to leverage server-side HTTP transport efficiences like 
connection pools, and so on.

To support API development that aggregates data from multiple sources, Xavi has a facility called a 
multi-backend adapter. A multi-backend adapter is a specialized plugin that allows an http.Handler to
have some context provided to allow multiple backends to be accessed. The backends have ServeHTTP mehods
that can be used to called the services assocaited with the backend, leveraging Xavi's server-side code
efficiencies, load balancers, and so on.

## How To Guide

Using multiple backends in a route for aggregation involves the following:

* Providing an implementation of a `plugin.MultiBackendHandlerFunc` that performs the aggregation using
multiple backends.
* Provide a factory function (type `plugin.MultiBackendAdapterFactory`) that can instantiate the `MultiBackendHandlerFunc`
implementation.
* Register the factory function via `plugin.RegisterMultiBackendAdapterFactory`

The `MultiBackendHandlerFunc` implementation is essentially an http.Handler function with an additional
argument of type `plugin.BackendHandlerMap`. The extra argument allows backend handlers to be looked up by 
name.

Here's an example of an implementation from `multiroute_server_test.go` in the service package:

<pre>

var bHandler plugin.MultiBackendHandlerFunc = func(m plugin.BackendHandlerMap, w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(bHandlerStuff))

    ah := m[backendA]
    ar := httptest.NewRecorder()
    ah.ServeHTTP(ar, r)
    assert.Equal(t, aBackendResponse, ar.Body.String())

    bh := m[backendB]
    br := httptest.NewRecorder()
    bh.ServeHTTP(br, r)
    assert.Equal(t, bBackendResponse, br.Body.String())
}
	
</pre>

The factory function from the test looks like this:

<pre>

var BMBAFactory = func(bhMap plugin.BackendHandlerMap) *plugin.MultiBackendAdapter {
    return &plugin.MultiBackendAdapter{
        Ctx:     bhMap,
        Handler: bHandler,
    }
}

</pre>

The test also registered the factory (see below). For applications built using the Xavi toolkit, the Run method
in the runner package has a plugin registration function argument. The multibackend adapter factory function should
be registered in the the implementation of the function passed to the run method.

<pre>

plugin.RegisterMultiBackendAdapterFactory(multiBackendAdapterFactory, BMBAFactory)

</pre>
