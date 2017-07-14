#!/bin/sh

set -ex

MAYA_API_SERVER_NETWORK=$1
NOMAD_ADDR=$2
NOMAD_CN_TYPE=$3
NOMAD_CN_NETWORK_CIDR=$4
NOMAD_CN_INTERFACE=$5
NOMAD_CS_PERSISTENCE_LOCATION=$6
NOMAD_CS_REPLICA_COUNT=$7

CONTAINER_IP_ADDR=$(ip -4 addr show scope global dev "${MAYA_API_SERVER_NETWORK}" | grep inet | awk '{print $2}' | cut -d / -f 1)

# Setup orch provider config OF M-APISERVER
# Here nomad running on global region's config file is considered

# nomad address related
#sed -i "s/__NOMAD_ADDR__/${NOMAD_ADDR}/g" /etc/mayaserver/orchprovider/nomad_global.INI

# cn related
#sed -i "s/__CN_TYPE__/${NOMAD_CN_TYPE}/g" /etc/mayaserver/orchprovider/nomad_global.INI
#sed -i "s/__CN_NETWORK_CIDR__/${NOMAD_CN_NETWORK_CIDR}/g" /etc/mayaserver/orchprovider/nomad_global.INI
#sed -i "s/__CN_INTERFACE__/${NOMAD_CN_INTERFACE}/g" /etc/mayaserver/orchprovider/nomad_global.INI

# cs related
#sed -i "s/__CS_PERSISTENCE_LOCATION__/${NOMAD_CS_PERSISTENCE_LOCATION}/g" /etc/mayaserver/orchprovider/nomad_global.INI
#sed -i "s/__CS_REPLICA_COUNT__/${NOMAD_CS_REPLICA_COUNT}/g" /etc/mayaserver/orchprovider/nomad_global.INI

# Start M-APISERVER service
exec /usr/local/bin/m-apiserver up -bind="${CONTAINER_IP_ADDR}" 1>&2

#exec /usr/local/bin/m-apiserver up 1>&2
