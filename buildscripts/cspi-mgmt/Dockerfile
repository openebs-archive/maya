#
# This Dockerfile builds a recent cspi-mgmt using the latest binary from
# cspi-mgmt  releases.
#

#openebs/cstor-base is the image that contains cstor related binaries and
#libraries - zpool, zfs, zrepl
#FROM openebs/cstor-base:ci
ARG BASE_IMAGE
FROM $BASE_IMAGE

COPY cspi-mgmt /usr/local/bin/
COPY entrypoint.sh /usr/local/bin/

RUN echo '#!/bin/bash\nif [ $# -lt 1 ]; then\n\techo "argument missing"\n\texit 1\nfi\neval "$*"\n' >> /usr/local/bin/execute.sh

RUN chmod +x /usr/local/bin/execute.sh
RUN apt install netcat -y
RUN chmod +x /usr/local/bin/entrypoint.sh

ARG ARCH
ARG DBUILD_DATE
ARG DBUILD_REPO_URL
ARG DBUILD_SITE_URL
LABEL org.label-schema.name="cspi-mgmt"
LABEL org.label-schema.description="CSPI Operator for OpenEBS cStor engine"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$DBUILD_DATE
LABEL org.label-schema.vcs-url=$DBUILD_REPO_URL
LABEL org.label-schema.url=$DBUILD_SITE_URL

ENTRYPOINT entrypoint.sh
EXPOSE 7676
