#
# This Dockerfile builds a recent cspc-operator using the latest binary from
# cspc-operator  releases.
#
FROM alpine:3.6

RUN apk add --no-cache \
    iproute2 \
    bash \
    curl \
    net-tools \
    mii-tool \
    procps \
    libc6-compat \
    ca-certificates

COPY cspc-operator /usr/local/bin/cspc-operator


ARG BUILD_DATE
LABEL org.label-schema.name="cspc-operator"
LABEL org.label-schema.description="CSPC Operator for OpenEBS cStor engine"
LABEL org.label-schema.url="http://www.openebs.io/"
LABEL org.label-schema.vcs-url="https://github.com/openebs/maya"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$BUILD_DATE

ENTRYPOINT ["/usr/local/bin/cspc-operator"]
