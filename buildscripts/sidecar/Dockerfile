FROM openebs/cstor:ci

ENV CSTOR_NETWORK="eth0"

RUN apt-get update && apt-get install -y \
    curl \
    open-iscsi \
    libaio1 \
    libaio-dev \
libkqueue-dev libssl1.0.0 rsyslog net-tools gdb apt-utils \
openssh-server vim sed sshpass

RUN apt -y install apt-file && apt-file update
RUN service ssh restart

COPY cstor-sidecar /usr/local/bin/
#COPY ZoL/cmd/.libs/zpool /usr/local/bin/
#COPY ZoL/cmd/.libs/zfs /usr/local/bin/
#COPY ZoL/lib/*.so* /usr/lib/
COPY entrypoint-sidecar.sh /usr/local/bin/

RUN chmod +x /usr/local/bin/entrypoint-sidecar.sh

ARG BUILD_DATE
LABEL org.label-schema.name="cstor"
LABEL org.label-schema.description="OpenEBS cstor"
LABEL org.label-schema.url="http://www.openebs.io/"
LABEL org.label-schema.vcs-url="https://github.com/openebs/cstor"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$BUILD_DATE

ENTRYPOINT entrypoint-sidecar.sh "${CSTOR_NETWORK}"
EXPOSE 7676
