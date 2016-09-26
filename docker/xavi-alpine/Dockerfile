# Start with the image that has mountebank installed
FROM mb-server-alpine

ENV HTTP_PROXY ${HTTP_PROXY}
ENV HTTPS_PROXY ${HTTPS_PROXY}
ENV http_proxy ${HTTP_PROXY}
ENV https_proxy ${HTTPS_PROXY}

ADD run-xavi.sh /opt/xtrac-xavi/
ADD xavi /usr/local/bin/

WORKDIR /opt/xtrac-xavi
EXPOSE 8080 9000
ENTRYPOINT ["/opt/xtrac-xavi/run-xavi.sh"]
