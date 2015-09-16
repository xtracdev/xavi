FROM mhart/alpine-node
MAINTAINER Doug Smith "doug.smith@fmr.com"
#ENV http_proxy http://<proxy server ip address>:<proxy server port>
#ENV https_proxy http://<proxy server ip address>:<proxy server port>

RUN npm install -g mountebank --production
RUN which mb

ENTRYPOINT [ "mb" ]
