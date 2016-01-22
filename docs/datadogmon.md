# Monitoring with DataDog

This document explains how to collect statsd and expvar metrics for use with [DataDog](https://www.datadoghq.com/).
Using DataDog with Xavi requires installing and configuring the DataDog agent, and configuring the Xavi environment
with configuration related to the monitoring set up.

## DataDog Agent Configuration

The DataDog site documents the basic installation of the agent (log into data dog then select Integrations > Agent
from the frame on the right hand side of the page).

Once the agent is installed, perform the following configuration.

### Set The Proxy Configuration

If your host requires a proxy to connect to the internet, edit `/opt/datadog-agent/etc/datadog.conf` and 
ensure both proxy_host and proxy_port are not commented out, and set them to the appropriate values.


### Monitoring Hostname

Also in datadog.conf, set host name to the host qualifier you might wish to use for filtering in datadog. Note this 
value can also be exported as an evironment variable. The environment variable value appears to be ignored if set in the agent.

### Configure DataDog for Statsd Collection

In datadog.conf, uncomment use_dogstatsd and set its value to yes. Uncomment dogstatsd_port and set it to 8125, or any other port
value.


### Golang Expvar Monitoring

In `/opt/datadog-agent/etc/conf.d`, rename go_expvar.yaml.example to go_expvar.yaml, and set the
hostname and port in expvar_url to the same values as the listener address for Xavi.

## Xavi Configuration

To configure Xavi for statsd and expvar monitoring, set the following environment variables:

* XAVI_DDHOST - Set the hostname in the DataDog statsd interface. Analog to hostname in datadog.conf. Note the the value in datadog.conf does not get overridden if a different value is set in XAVI_DDHOST.
* XAVI_STATSD_NAMESPACE - String prepended to metric names emitted from Xavi.
* XAVI_USE_DATADOG_STATSD - set to 1 when sending statsd data to DataDog
* XAVI_STATSD_ADDRESS - set to the hostname:port of the statsd agent. If using DataDog statsd, the port component
of this variable should be set to the same value as dogstatsd_port as described above.

Finally, note that when running Xavi in listener mode, the expvar_url set in the DataDig go_expvar.yaml file.
