FROM mhart/alpine-node
MAINTAINER Doug Smith "doug.smith@***REMOVED***"
ENV http_proxy http://***REMOVED***
ENV https_proxy http://***REMOVED***

RUN npm install -g mountebank --production
RUN which mb

ENTRYPOINT [ "mb" ]
