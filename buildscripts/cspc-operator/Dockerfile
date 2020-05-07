#
# This Dockerfile builds a recent cspc-operator using the latest binary from
# cspc-operator  releases.
#
FROM alpine:3.11.5

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

ARG ARCH
ARG DBUILD_DATE
ARG DBUILD_REPO_URL
ARG DBUILD_SITE_URL
LABEL org.label-schema.name="cspc-operator"
LABEL org.label-schema.description="CSPC Operator for OpenEBS cStor engine"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$DBUILD_DATE
LABEL org.label-schema.vcs-url=$DBUILD_REPO_URL
LABEL org.label-schema.url=$DBUILD_SITE_URL

ENTRYPOINT ["/usr/local/bin/cspc-operator"]
