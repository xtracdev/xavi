# Start with the image that has mountebank installed
FROM mb-server-alpine
MAINTAINER Doug Smith "doug.smith@fmr.com"
#ENV http_proxy http://<proxy server ip address>:<proxy server port>
#ENV https_proxy http://<proxy server ip address>:<proxy server port>

ADD run-xavi.sh /opt/xtrac-xavi/
ADD xavi /usr/local/bin/

WORKDIR /opt/xtrac-xavi
EXPOSE 8080 9000
ENTRYPOINT ["/opt/xtrac-xavi/run-xavi.sh"]
