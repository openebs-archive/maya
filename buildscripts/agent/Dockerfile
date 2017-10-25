#
# This Dockerfile builds a recent maya agent using the latest binary from
# maya-agent  releases.
#

FROM alpine:3.6

# TODO: The following env variables should be auto detected.
ENV MAYA_AGENT_NETWORK="eth0"

RUN apk add --no-cache \
    iproute2 \
    curl \
    net-tools \
    mii-tool \
    procps \
    libc6-compat

COPY maya-agent /usr/local/bin/
COPY entrypoint.sh /usr/local/bin/

RUN chmod +x /usr/local/bin/entrypoint.sh

ARG BUILD_DATE
LABEL org.label-schema.name="m-agent"
LABEL org.label-schema.description="OpenEBS Agent"
LABEL org.label-schema.url="http://www.openebs.io/"
LABEL org.label-schema.vcs-url="https://github.com/openebs/maya"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$BUILD_DATE

ENTRYPOINT entrypoint.sh "${MAYA_AGENT_NETWORK}"
EXPOSE 7676
