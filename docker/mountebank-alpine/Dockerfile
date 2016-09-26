FROM mhart/alpine-node

ENV HTTP_PROXY ${HTTP_PROXY}
ENV HTTPS_PROXY ${HTTPS_PROXY}
ENV http_proxy ${HTTP_PROXY}
ENV https_proxy ${HTTPS_PROXY}

RUN npm install -g mountebank --production
RUN which mb

ENTRYPOINT [ "mb", "--mock" ]
