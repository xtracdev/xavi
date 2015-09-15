# Start with the image that has mountebank installed
FROM mb-server-alpine
MAINTAINER Doug Smith "doug.smith@***REMOVED***"
ENV http_proxy http://***REMOVED***
ENV https_proxy http://***REMOVED***

ADD run-xavi.sh /opt/xtrac-xavi/
ADD xavi /usr/local/bin/

WORKDIR /opt/xtrac-xavi
EXPOSE 8080 9000
ENTRYPOINT ["/opt/xtrac-xavi/run-xavi.sh"]
