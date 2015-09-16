## Misc Notes 

These will be cleaned up and enhanced, then moved to the README

### Vagrant Image

A vagrant box for the Virtualbox provider is available to team members via the Hashicorp Atlas
repository. Downloading and booting the box provides standard environment components, such as
docker, consul, graphite, statsd, elastisearch, logstash, kibana, fluentd, etc.

The details on how the box was created are available in the [xavi-docker](https://github.com/xtracdev/xavi-docker)
project.

When running xavi locally using services running in the Vagrant box, set environment variables
as needed. Current the consul and statsd directories must be set:

<pre>
export XAVI_CONSUL_AGENT_IP=172.20.20.70
export XAVI_STATSD_ADDRESS=172.20.20.70:8125
</pre>

A pre-built box (private image) is available via Hashicorp Atlas to team members.

### Codeship Setup - Copy to S3 Bucket

Note for simple CI on commit there's no need to do an S3 deploy

Test setup:

<pre>
go get github.com/tools/godep
</pre>

Test commands

<pre>
godep go test ./...
</pre>

1st deploy
* Custom Script

		godep go build

2nd deploy
* Amazon S3

		access key id, secret access key
		region - us-west-1
		localpath - xavi (go build artifact)
		s3 bucket - ds-codeship
		acl - bucket-owner-full-control


### Register Service - Consul

/v1/agent/service/register
<pre>
{
  "ID": "demo-service-1",
  "Name": "demo-service",
  "Tags": [
    "demosvc",
    "v1"
  ],
  "Address": "172.20.20.70",
  "Port": 3000
}
</pre>

<pre>
{
  "ID": "demo-service-2",
  "Name": "demo-service",
  "Tags": [
    "demosvc",
    "v1"
  ],
  "Address": "172.20.20.70",
  "Port": 3100
}
</pre>




### Register Health Check for Service

/v1/agent/check/register
<pre>
{
  "ID": "demo-service-check-1",
  "service_id": "demo-service-1",
  "service-name":"demo-service",
  "Name": "hello",
  "Notes": "Get /hello",
  "HTTP": "http://172.20.20.70:3000/hello",
  "Interval": "10s"
}
</pre>

<pre>
{
  "ID": "demo-service-check-2",
  "service_id": "demo-service-2",
  "service-name":"demo-service",
  "Name": "hello",
  "Notes": "Get /hello",
  "HTTP": "http://172.20.20.70:3100/hello",
  "Interval": "10s"
}
</pre>

### Registered Service DNS query

    dig @172.20.20.70 -p 8600 demo-service.service.consul SRV
    dig @172.20.20.70 -p 8600 v1.demo-service.service.consul SRV

### Deregister Health Check

	curl http://172.20.20.70:8500/v1/agent/check/deregister/demo-service-check

### Deregister service

	curl http://172.20.20.70:8500/v1/agent/service/deregister/demo-service




### Expvar URI on listener host:port

		/debug/vars

### Log Rotation

In /etc/logrotate.d, sudo vi xavi-demo-svc.log:

		/home/vagrant/xavi-logs/demo.log {
			missingok
			copytruncate
			size 50k
			create 755 vagrant vagrant
			su vagrant vagrant
			rotate 20
		}

Then add an entry in /etc/crontab (sudo vi /etc/crontab)

		30 *	* * * 	root	/usr/sbin/logrotate /etc/logrotate.d/xavi-demo-svc.log

Or, alternatively create a script in /etc/cron.hourly (make sure to chmod +x the script)

		#!/bin/sh
		/usr/sbin/logrotate /etc/logrotate.d/xavi-demo-svc.log

logrotate is fussy about the permissions. I created a umask of 022 in the .vagrant .bashrc


The above assumes writing the stdout to a logfile named xavi-demo-svc.log

		xavi listen -ln demo-listener -address 0.0.0.0:11223 >> xavi-logs/demo.log

Note if you don't 'append' via >> you will not see the file size diminish
on truncation (it will have null bytes prepended to the up to the last write offset).

### Statsd Setup - Ubuntu

		apt-get install nodejs npm  
		sudo apt-get install nodejs npm  
		sudo apt-get install git
		git clone https://github.com/etsy/statsd/
		cd statsd/
		cp exampleConfig.js config.js
		vi config.js
		nodejs stats.js ./config.js

config.js contents:

<pre>
{
graphitePort: 2003
, graphiteHost: "172.20.20.50"
,port: 8125
, backends: [ "./backends/console","./backends/graphite" ]
, debug: true
, dumpMessages: true
}
</pre>



