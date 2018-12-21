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

ARG BUILD_DATE
LABEL org.label-schema.name="cstor-volume-mgmt"
LABEL org.label-schema.description="OpenEBS"
LABEL org.label-schema.url="http://www.openebs.io/"
LABEL org.label-schema.vcs-url="https://github.com/openebs/maya"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$BUILD_DATE

ENTRYPOINT entrypoint.sh
EXPOSE 7676 7777
