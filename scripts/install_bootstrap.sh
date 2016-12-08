#!/bin/bash

set -ex

CURDIR=`pwd`

mkdir -p /etc/maya.d/scripts

sudo chmod a+w /etc/maya.d

cd /etc/maya.d/scripts

echo Fetching consul script ...
#wget -q https://raw.githubusercontent.com/openebs/maya/master/scripts/install_consul.sh -O install_consul.sh
curl -sSL https://raw.githubusercontent.com/openebs/maya/master/scripts/install_consul.sh -o install_consul.sh

cd ${CURDIR}
