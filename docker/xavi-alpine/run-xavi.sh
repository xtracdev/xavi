#!/bin/sh

# Start the co-hosted mountebank server. We need one in the same container to
# allow testing of the pref-local load balancing policy.
mb&

# Boot XAVI in rest agent mode
export XAVI_KVSTORE_URL=file:///opt/xtrac-xavi/xavikeystore

#cat /opt/xtrac-xavi/xavikeystore

# Some lines to give insight into Docker context
#ping -c 5 mbhost
xavi boot-rest-agent -address 0.0.0.0:8080
