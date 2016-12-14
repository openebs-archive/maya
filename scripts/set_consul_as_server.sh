#!/bin/bash

set -e

if [ $# -ne 4 ]; then
    echo usage: $0 SELF_IPV4 SELF_HOSTNAME ALL_SERVERS_IPV4 S_NODES
    exit 1
fi

SELF_IPV4=$1
SELF_HOSTNAME=$2
ALL_SERVERS_IPV4=$3
S_NODES=$4

echo "Set consul in server mode ..."

# Place consul config template 
sudo cp /etc/maya.d/templates/consul-server.json.tmpl /etc/consul.d/server/consul-server.json

# Place systemd service template for consul
sudo cp /etc/maya.d/templates/consul-server.service.tmpl /etc/systemd/system/consul-server.service

# Replace the placeholders with actual values
sudo sed -e "s|__SELF_HOSTNAME__|$SELF_HOSTNAME|g" -i /etc/consul.d/server/consul-server.json
sudo sed -e "s|__SELF_IPV4__|$SELF_IPV4|g" -i /etc/consul.d/server/consul-server.json
sudo sed -e "s|__S_NODES__|$S_NODES|g" -i /etc/consul.d/server/consul-server.json
sudo sed -e "s|__ALL_SERVERS_IPV4__|$ALL_SERVERS_IPV4|g" -i /etc/consul.d/server/consul-server.json
