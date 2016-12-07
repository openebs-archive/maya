#!/bin/bash

set -ex

CURDIR=`pwd`

mkdir -p /etc/maya.d/scripts

sudo chmod a+w /etc/maya.d

cd /etc/maya.d/scripts

echo Fetching consul script ...
wget -q https://github.com/openebs/maya/blob/master/scripts/install_consul.sh

cd ${CURDIR}
