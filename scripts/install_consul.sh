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

# Location to hold consul's config files
# Remove if already present
# NOTE: this is install only script
sudo rm -rf /etc/consul.d/

sudo mkdir -p /etc/consul.d/bootstrap
sudo mkdir -p /etc/consul.d/server
sudo mkdir -p /etc/consul.d/client

sudo chmod a+w /etc/consul.d
sudo chmod a+w /etc/consul.d/bootstrap
sudo chmod a+w /etc/consul.d/server
sudo chmod a+w /etc/consul.d/client

# Location to store consul's persistent data between reboots
# Remove if already present
# NOTE: this is install only script
sudo rm -rf /opt/consul/

sudo mkdir -p /opt/consul/data
sudo chmod a+w /opt/consul/data

cd ${CURDIR}
