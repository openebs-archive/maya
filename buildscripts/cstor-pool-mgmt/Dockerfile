#
# This Dockerfile builds a recent cstor-pool-mgmt using the latest binary from
# cstor-pool-mgmt  releases.
#

#openebs/cstor-base is the image that contains cstor related binaries and
#libraries - zpool, zfs, zrepl
#FROM openebs/cstor-base:ci
ARG BASE_IMAGE
FROM $BASE_IMAGE

COPY cstor-pool-mgmt /usr/local/bin/
COPY entrypoint.sh /usr/local/bin/

RUN echo '#!/bin/bash\nif [ $# -lt 1 ]; then\n\techo "argument missing"\n\texit 1\nfi\neval "$*"\n' >> /usr/local/bin/execute.sh

RUN chmod +x /usr/local/bin/execute.sh
RUN apt install netcat -y
RUN chmod +x /usr/local/bin/entrypoint.sh

ARG BUILD_DATE
LABEL org.label-schema.name="cstor-pool-mgmt"
LABEL org.label-schema.description="OpenEBS"
LABEL org.label-schema.url="http://www.openebs.io/"
LABEL org.label-schema.vcs-url="https://github.com/openebs/maya"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$BUILD_DATE

ENTRYPOINT entrypoint.sh
EXPOSE 7676
