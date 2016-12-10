#!/bin/bash

set -e

CURDIR=`pwd`

# Remove if already present
# NOTE: this is install only script
echo "Cleaning old maya boostrapping if any"
sudo rm -rf /etc/maya.d/

sudo mkdir -p /etc/maya.d/scripts
sudo mkdir -p /etc/maya.d/templates

sudo chmod a+w /etc/maya.d/
sudo chmod a+w /etc/maya.d/scripts
sudo chmod a+w /etc/maya.d/templates

# Fetch various install scripts
cd /etc/maya.d/scripts
echo "Fetching consul installer scripts ..."

curl -sSL https://raw.githubusercontent.com/openebs/maya/master/scripts/install_consul.sh -o install_consul.sh > /dev/null 2>&1
curl -sSL https://raw.githubusercontent.com/openebs/maya/master/scripts/set_consul_as_server.sh -o set_consul_as_server.sh > /dev/null 2>&1
curl -sSL https://raw.githubusercontent.com/openebs/maya/master/scripts/get_first_private_ip.sh -o get_first_private_ip.sh > /dev/null 2>&1
curl -sSL https://raw.githubusercontent.com/openebs/maya/master/scripts/start_consul_server.sh -o start_consul_server.sh > /dev/null 2>&1

# Fetch various templates
cd /etc/maya.d/templates
echo "Fetching consul server config templates ..."

curl -sSL https://raw.githubusercontent.com/openebs/maya/master/templates/consul-server.json.tmpl -o consul-server.json.tmpl > /dev/null 2>&1
curl -sSL https://raw.githubusercontent.com/openebs/maya/master/templates/consul-server.service.tmpl -o consul-server.service.tmpl > /dev/null 2>&1

cd ${CURDIR}
