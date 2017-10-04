#!/bin/bash

set -e

ETCD_VER="v3.0.15"
ETCD_DOWNLOAD_URL=https://github.com/coreos/etcd/releases/download
CURDIR=`pwd`

# Remove if already present
# NOTE: this is install only script
echo "Cleaning old etcd installation if any"
sudo rm -rf /usr/bin/etcd
sudo rm -rf /usr/bin/etcdctl
sudo rm -rf /var/lib/etcd

cd /tmp/

if [ ! -f "./etcd-${ETCD_VER}-linux-amd64.tar.gz" ]; then
  echo "Fetching etcd ${ETCD_VER} ..."
  curl -sSL ${ETCD_DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o etcd-${ETCD_VER}-linux-amd64.tar.gz  
fi

echo "Installing etcd ${ETCD_VER} ..."
mkdir -p etcdall && tar xzvf etcd-${ETCD_VER}-linux-amd64.tar.gz -C etcdall --strip-components=1
sudo chmod +x etcdall/etcd
sudo chmod +x etcdall/etcdctl
sudo mv etcdall/etcd /usr/bin/etcd
sudo mv etcdall/etcdctl /usr/bin/etcdctl

# Setup data directory for etcd
sudo mkdir -p /var/lib/etcd
sudo chmod a+w /var/lib/etcd

cd ${CURDIR}
