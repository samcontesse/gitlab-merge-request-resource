ARG ALPINE_VERSION=3.16
FROM alpine:$ALPINE_VERSION
RUN apk add --update-cache git openssh-client \
 && git --version
COPY check /opt/resource/
COPY in /opt/resource/
COPY out /opt/resource/
