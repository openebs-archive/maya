#!/bin/sh

set -ex

MAYA_NODEBOT_NETWORK=$1

CONTAINER_IP_ADDR=$(ip -4 addr show scope global dev "${MAYA_NODEBOT_NETWORK}" | grep inet | awk '{print $2}' | cut -d / -f 1)

# Start nodebot service
exec /usr/local/bin/maya-nodebot up -bind="${CONTAINER_IP_ADDR}" 1>&2
