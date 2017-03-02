#!/bin/bash

set -e

MAYA_VERSION="0.0.1"
CURDIR=`pwd`

# Remove if already present
echo "Cleaning old Mayaserver installation if any"
sudo rm -rf /usr/bin/mayaserver

cd /tmp/

echo "Fetching Mayaserver ${MAYA_VERSION} ..."
wget -q https://github.com/openebs/mayaserver/releases/download/${MAYA_VERSION}/mayaserver-linux_amd64.zip -o mayaserver.zip

echo "Installing Mayaserver ${MAYA_VERSION} ..."
unzip mayaserver.zip
sudo chmod +x mayaserver
sudo mv mayaserver /usr/bin/mayaserver

mayaserver version

cd ${CURDIR}
