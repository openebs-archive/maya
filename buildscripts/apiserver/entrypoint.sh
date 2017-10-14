#!/bin/sh

set -ex

MAYA_API_SERVER_NETWORK=$1

CONTAINER_IP_ADDR=$(ip -4 addr show scope global dev "${MAYA_API_SERVER_NETWORK}" | grep inet | awk '{print $2}' | cut -d / -f 1)

# Start apiserver service
exec /usr/local/bin/maya-apiserver up -bind="${CONTAINER_IP_ADDR}" 1>&2
