#!/bin/bash

set -ex

CURDIR=`pwd`

sudo mkdir -p /etc/maya.d/scripts

sudo chmod a+w /etc/maya.d/

# Fetch various install scripts
sudo chmod a+w /etc/maya.d/scripts
cd /etc/maya.d/scripts

echo Fetching consul script ...
curl -sSL https://raw.githubusercontent.com/openebs/maya/master/scripts/install_consul.sh -o install_consul.sh

cd ${CURDIR}
