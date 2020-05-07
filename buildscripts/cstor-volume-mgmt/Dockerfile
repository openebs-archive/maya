#
# This Dockerfile builds a recent cstor-volume-mgmt using the latest binary from
# cstor-volume-mgmt  releases.
#

FROM ubuntu:16.04
RUN apt-get update; exit 0
RUN apt-get -y install rsyslog

RUN mkdir -p /usr/local/etc/istgt

COPY cstor-volume-mgmt /usr/local/bin/
COPY entrypoint.sh /usr/local/bin/

RUN chmod +x /usr/local/bin/entrypoint.sh

ARG ARCH
ARG DBUILD_DATE
ARG DBUILD_REPO_URL
ARG DBUILD_SITE_URL
LABEL org.label-schema.name="cstor-volume-mgmt"
LABEL org.label-schema.description="OpenEBS cStor Volume Operator"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$DBUILD_DATE
LABEL org.label-schema.vcs-url=$DBUILD_REPO_URL
LABEL org.label-schema.url=$DBUILD_SITE_URL

ENTRYPOINT entrypoint.sh
EXPOSE 7676 7777
