## XAVI - The XTRAC API Platform

[![Codeship Status for xtracdev/xavi](https://codeship.com/projects/fa2d0050-3e24-0133-f8eb-5a0949beaeb8/status?branch=master)](https://codeship.com/projects/102711)


Xavi is a software layer that decouples API consumers from the systems that provide the 
underlying capabilities of the APIs, and allows an API contract to be defined and maintained from 
the consumer perspective, implemented as a facade in front of the services and applications providing 
API functionality.

Xavi provides several key features:

* **HTTP Reverse Proxy.** HTTP endpoints can be exposed by the XTRAC API gateway to
API consumers, with the gateway handling the routing to the servers providing
API services.
* **Load balancing.** The XTRAC API Gateway provides the ability to load balance
among multiple servers providing capabilities used in an API.
* **Plugin Mechanism.** The XTRAC API Gateway provides a well defined plugin interface
that allows implementing the decorator pattern on HTTP calls, enabling things like
message transformation, url rewriting, and protocol translation.
* **Configuration via Command Line and REST Services.** The configuration used
at runtime by the XTRAC API Gateway can be configured both at the command line and
via a REST service API.
* **Pluggable configuration store.** Supported stores currently include Consul and a memory-based store that can be flushed to disk.

### Contributing

To contribute, you must certify you agree with the [Developer Certificate of Origin](http://developercertificate.org/)
by signing your commits via `git -s`. To create a signature, configure your user name and email address in git.
Sign with your real name, do not use pseudonyms or submit anonymous commits.


In terms of workflow:

0. For significant changes or improvement, create an issue before commencing work.
1. Fork the respository, and create a branch for your edits.
2. Add tests that cover your changes, unit tests for smaller changes, acceptance test
for more significant functionality.
3. Run gofmt on each file you change before committing your changes.
4. Run golint on each file you change before committing your changes.
5. Make sure all the tests pass before committing your changes.
6. Commit your changes and issue a pull request.

### Quick Start

A guide for trying out xavi is available in the xavi github repository:

[Quick Start Guide](docs/quickstart.md)

### Developing Xavi

Developer notes for building and coding xavi are available in the xavi github repository:

[Developer Notes](docs/developing.md)

### License

(c) 2015 Fidelity Investments
Licensed under the Apache License, Version 2.0






