#!/bin/sh

set -ex

MAYA_AGENT_NETWORK=$1

CONTAINER_IP_ADDR=$(ip -4 addr show scope global dev "${MAYA_AGENT_NETWORK}" | grep inet | awk '{print $2}' | cut -d / -f 1)

# Start agent service
exec /usr/local/bin/maya-agent up -bind="${CONTAINER_IP_ADDR}" 1>&2
