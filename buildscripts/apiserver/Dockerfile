#
# This Dockerfile builds a recent maya api server using the latest binary from
# maya api server's releases.
#

FROM alpine:3.11.5

# TODO: The following env variables should be auto detected.
ENV MAYA_API_SERVER_NETWORK="eth0"

RUN apk add --no-cache \
    iproute2 \
    bash \
    curl \
    net-tools \
    mii-tool \
    procps \
    libc6-compat \
    ca-certificates
RUN mkdir -p /etc/apiserver/orchprovider
RUN mkdir -p /etc/apiserver/specs

COPY demo-vol1.yaml /etc/apiserver/specs/
COPY maya-apiserver /usr/local/bin/
#COPY mayactl /usr/local/bin/

COPY entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh

ARG ARCH
ARG DBUILD_DATE
ARG DBUILD_REPO_URL
ARG DBUILD_SITE_URL
LABEL org.label-schema.name="m-apiserver"
LABEL org.label-schema.description="API server for OpenEBS"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$DBUILD_DATE
LABEL org.label-schema.vcs-url=$DBUILD_REPO_URL
LABEL org.label-schema.url=$DBUILD_SITE_URL

ENTRYPOINT entrypoint.sh "${MAYA_API_SERVER_NETWORK}"

EXPOSE 5656
