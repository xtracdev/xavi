Originally I was looking to put both consul and xavi in the same docker container. Since in a docker container 
you can have a single CMD/ENTRY, I thought I would install consul as a process started via upstart/init,
but it turns out the Docker base image for ubuntu is extremely stripped down and doesn't offer this.

I will likely build a vagrant image to get around this limitation, plus have something I can leverage 
both from a CD pipeline in Amazon as well as something that can be passed around and ran internally.

Below is the original Dockerfile contents and supporting files which come in handy for the 
Vagrant config.

Note this Dockerfile was an intermediate representation that went from the base ubuntu
version to one that attempted to include upstart, etc. This never ran consul sucessfully.

Dockerfile:

		FROM phusion/baseimage:0.9.16
		MAINTAINER Doug Smith "doug.smith.mail@gmail.com"
		ENV http_proxy http://***REMOVED***:8000
		ENV https_proxy http://***REMOVED***:8000
		
		RUN apt-get -yqq update
		RUN apt-get install -y zip
		RUN apt-get install -y wget
		WORKDIR /usr/local/bin
		RUN wget https://dl.bintray.com/mitchellh/consul/0.4.1_linux_amd64.zip
		RUN unzip 0.4.1_linux_amd64.zip
		#RUN mkdir -p /var/consul
		#RUN mkdir -p /etc/consul.d/server
		#COPY consul.atest.json /etc/consul.d/server/config.json
		#COPY consul.test.conf /etc/init/consul.conf
		#RUN start consul
		#EXPOSE 8500
		#RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
		RUN mkdir /etc/service/consul
		ADD runconsul.sh /etc/service/consul/run
		
		CMD ["/sbin/my_init"]

consul.atest.json

		{
		    "bootstrap": true,
		    "server": true,
		    "datacenter": "slc1",
		    "data_dir": "/var/consul",
		    "log_level": "INFO",
		    "enable_syslog": true
		}

consul.test.conf:

		description "Consul server process"
		
		start on (local-filesystems and net-device-up IFACE=eth0)
		stop on runlevel [!12345]
		
		respawn
		
		
		exec consul agent -config-dir /etc/consul.d/server

runconsul.sh:

		#!/bin/sh
		exec /usr/local/bin/consul agent -server -bootstrap-expect 1 -data-dir /var/consul -node=agent-one >>/var/log/consul.log 2>&1
