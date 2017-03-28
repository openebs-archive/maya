#!/bin/bash

set -e

if [ $# -ne 2 ]; then
    echo usage: $0 SELF_IPV4 ALL_SERVERS_IPV4_N_PORTS
    exit 1
fi

SELF_IPV4=$1
ALL_SERVERS_IPV4_N_PORTS=$2

echo "Set nomad in client mode ..."

# Place nomad config template 
sudo cp /etc/maya.d/templates/nomad-client.hcl.tmpl /etc/nomad.d/client/nomad-client.hcl

# Place systemd service template for nomad
sudo cp /etc/maya.d/templates/nomad-client.service.tmpl /etc/systemd/system/nomad-client.service

# Replace the placeholders with actual values
sudo sed -e "s|__SELF_IPV4__|$SELF_IPV4|g" -i /etc/nomad.d/client/nomad-client.hcl
sudo sed -e "s|__ALL_SERVERS_IPV4_N_PORTS__|$ALL_SERVERS_IPV4_N_PORTS|g" -i /etc/nomad.d/client/nomad-client.hcl

# Set the env variable to bind nomad cli to self ip
grep "export NOMAD_ADDR=http://${SELF_IPV4}:4646" ~/.profile || \
  echo "export NOMAD_ADDR=http://${SELF_IPV4}:4646" >> ~/.profile
export NOMAD_ADDR=http://${SELF_IPV4}:4646

# set the Maya apiserver env variable
grep "export MAPI_ADDR=http://${SELF_IPV4}:5656" ~/.profile || \
  echo "export MAPI_ADDR=http://${SELF_IPV4}:5656" >> ~/.profile
export MAPI_ADDR=http://${SELF_IPV4}:5656
