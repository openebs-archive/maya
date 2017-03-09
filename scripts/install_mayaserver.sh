#!/bin/bash

set -e

MAYA_VERSION="0.0.3"
CURDIR=`pwd`

# Remove if already present
# echo "Cleaning old Mayaserver installation if any"
#sudo rm -rf /usr/bin/mayaserver

if [[ $(which mayaserver >/dev/null && mayaserver version | head -n 1 | cut -d ' ' -f 2) == "v$MAYA_VERSION" ]]; then
    echo "Mayaserver v$MAYA_VERSION already installed; Skipping"
    exit
fi

cd /tmp/

if [ ! -f "./mayaserver_${MAYA_VERSION}.zip" ]; then
echo "Fetching Mayaserver ${MAYA_VERSION} ..."
curl -sSL https://github.com/openebs/mayaserver/releases/download/${MAYA_VERSION}/mayaserver-linux_amd64.zip -o mayaserver.zip
fi

echo "Installing Mayaserver ${MAYA_VERSION} ..."
unzip mayaserver.zip
sudo chmod +x mayaserver
sudo mv mayaserver /usr/bin/mayaserver

cd ${CURDIR}
