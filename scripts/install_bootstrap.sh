#!/bin/bash

set -e

CURDIR=`pwd`

# Remove if already present
# NOTE: this is install only script
sudo rm -rf /etc/maya.d/

sudo mkdir -p /etc/maya.d/scripts
sudo mkdir -p /etc/maya.d/templates

sudo chmod a+w /etc/maya.d/
sudo chmod a+w /etc/maya.d/scripts
sudo chmod a+w /etc/maya.d/templates

# Fetch various install scripts
cd /etc/maya.d/scripts
echo Fetching consul installer scripts ...

curl -sSL https://raw.githubusercontent.com/openebs/maya/master/scripts/install_consul.sh -o install_consul.sh
curl -sSL https://raw.githubusercontent.com/openebs/maya/master/scripts/set_consul_as_server.sh -o set_consul_as_server.sh
curl -sSL https://raw.githubusercontent.com/openebs/maya/master/scripts/get_first_private_ip.sh -o get_first_private_ip.sh

# Fetch various templates
cd /etc/maya.d/templates
echo Fetching consul server config templates ...

curl -sSL https://raw.githubusercontent.com/openebs/maya/master/templates/consul-server.json.tmpl -o consul-server.json.tmpl
curl -sSL https://raw.githubusercontent.com/openebs/maya/master/templates/consul-server-init.conf.tmpl -o consul-server-init.conf.tmpl

cd ${CURDIR}
