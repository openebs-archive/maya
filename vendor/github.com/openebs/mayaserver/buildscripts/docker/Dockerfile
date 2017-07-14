#
# This Dockerfile builds a recent maya api server using the latest binary from
# maya api server's releases.
#
# This Dockerfile may be run with proper ENV pairs
# e.g. docker run --env NOMAD_ADDR=http://172.28.128.3:4646
#

FROM ubuntu:16.04

# TODO: The following env variables should be auto detected. 
ENV MAYA_API_SERVER_VERSION="0.2-RC4" \
    MAYA_API_SERVER_NETWORK="eth0" \
    TERM="vt100" \
    NOMAD_ADDR="http://172.28.128.3:4646" \
    NOMAD_CN_TYPE="host" \
    NOMAD_CN_NETWORK_CIDR="172.28.128.1/24" \
    NOMAD_CN_INTERFACE="enp0s8" \
    NOMAD_CS_PERSISTENCE_LOCATION="/tmp/" \
    NOMAD_CS_REPLICA_COUNT="2"
    
RUN apt-get update && apt-get install -y \
    iproute2 \
    curl \
    net-tools \
    watch \
 && rm -rf /var/lib/apt/lists/*
    
RUN mkdir -p /etc/mayaserver/orchprovider
COPY nomad_global.INI.tmpl /etc/mayaserver/orchprovider/nomad_global.INI
COPY demo-vol1.yaml /etc/mayaserver/specs/
COPY m-apiserver maya /usr/local/bin/

COPY entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT entrypoint.sh \
  "${MAYA_API_SERVER_NETWORK}" \
  "${NOMAD_ADDR}" \
  "${NOMAD_CN_TYPE}" \
  "${NOMAD_CN_NETWORK_CIDR}" \
  "${NOMAD_CN_INTERFACE}" \
  "${NOMAD_CS_PERSISTENCE_LOCATION}" \
  "${NOMAD_CS_REPLICA_COUNT}"

EXPOSE 5656
