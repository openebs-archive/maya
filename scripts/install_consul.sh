#!/bin/bash

set -ex

CONSUL_VERSION="0.7.0"
CURDIR=`pwd`

echo Fetching Consul ${CONSUL_VERSION} ...

cd /tmp/

curl -sSL https://releases.hashicorp.com/consul/${CONSUL_VERSION}/consul_${CONSUL_VERSION}_linux_amd64.zip -o consul.zip
echo Installing Consul...
unzip consul.zip
sudo chmod +x consul
sudo mv consul /usr/bin/consul

sudo mkdir -p /etc/consul.d
sudo chmod a+w /etc/consul.d

# This is very important as this script may be used in some workflow
# The directory of the workflow should not be modified by individual scripts
cd ${CURDIR}
