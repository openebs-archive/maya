#!/bin/bash

set -e
releaseurl="https://api.github.com/repos/openebs/mayaserver/releases"
CURDIR=`pwd`

if [[ $(which mayaserver >/dev/null && mayaserver version | head -n 1 | cut -d ' ' -f 2 | sed 's/dev//' | cut -d "'" -f 2) == "$MAYA_VERSION" ]]; then    
echo "Mayaserver v$MAYA_VERSION already installed; Skipping"
    exit
fi

cd /tmp/

if [ ! -f "./mayaserver_${MAYA_VERSION}.zip" ]; then
echo "Fetching Mayaserver ${MAYA_VERSION} ..."
#curl -sSL https://github.com/openebs/mayaserver/releases/download/${MAYA_VERSION}/mayaserver-linux_amd64.zip -o mayaserver.zip
wget $(curl -sS $releaseurl | grep "browser_download_url" | awk '{print $2}' | tr -d '"' | head -n 2 | tail -n 1)
exit
fi
echo "Installing Mayaserver ${MAYA_VERSION} ..."
unzip mayaserver-linux_amd64.zip
sudo chmod +x mayaserver
sudo mv mayaserver /usr/bin/mayaserver

# Setup config directory for mayaserver
sudo mkdir -p /etc/mayaserver/orchprovider
sudo chmod a+w /etc/mayaserver/orchprovider


cd ${CURDIR}
