#!/bin/bash

set -e

if [ $# -ne 2 ]; then
    echo usage: $0 SELF_IPV4 ALL_SERVERS_IPV4
    exit 1
fi

SELF_IPV4=$1
ALL_SERVERS_IPV4=$2

echo "Set consul in client mode ..."

# Place consul config template 
sudo cp /etc/maya.d/templates/consul-client.json.tmpl /etc/consul.d/client/consul-client.json

# Place systemd service template for consul
sudo cp /etc/maya.d/templates/consul-client.service.tmpl /etc/systemd/system/consul-client.service

# Replace the placeholders with actual values
sudo sed -e "s|__SELF_IPV4__|$SELF_IPV4|g" -i /etc/consul.d/client/consul-client.json
sudo sed -e "s|__ALL_SERVERS_IPV4__|$ALL_SERVERS_IPV4|g" -i /etc/consul.d/client/consul-client.json
