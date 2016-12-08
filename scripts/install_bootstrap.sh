#!/bin/bash

set -ex

CURDIR=`pwd`

sudo mkdir -p /etc/maya.d/{scripts,templates}

sudo chmod a+w /etc/maya.d/
sudo chmod a+w /etc/maya.d/scripts
sudo chmod a+w /etc/maya.d/templates

# Fetch various install scripts
cd /etc/maya.d/scripts
echo Fetching consul installer script ...
curl -sSL https://raw.githubusercontent.com/openebs/maya/master/scripts/install_consul.sh -o install_consul.sh

# Fetch various templates
cd /etc/maya.d/templates
echo Fetching consul server config template ...
curl -sSL https://raw.githubusercontent.com/openebs/maya/master/templates/consul-server.json.tmpl -o consul-server.json.tmpl

cd ${CURDIR}
