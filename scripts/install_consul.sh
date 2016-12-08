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
sudo mkdir -p /etc/consul.d/{bootstrap,server,client}
sudo chmod -R a+w /etc/consul.d

# Location to store consul's persistent data between reboots
sudo mkdir -p /opt/consul/data

cd ${CURDIR}
