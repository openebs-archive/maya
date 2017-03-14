#!/bin/bash

echo "Provisioning network on master"
set -e

FLANNEL_VER="v0.6.2"
FLANNEL_DOWNLOAD_URL=https://github.com/coreos/flannel/releases/download
CURDIR=`pwd`

cd /tmp/

if [ ! -f "./tmp/flannel-${FLANNEL_VER}-linux-amd64.tar.gz" ]; then
  echo "Fetching flannel ${FLANNEL_VER} ..."
  curl -sSL ${FLANNEL_DOWNLOAD_URL}/${FLANNEL_VER}/flannel-${FLANNEL_VER}-linux-amd64.tar.gz -o flannel-${FLANNEL_VER}-linux-amd64.tar.gz
fi

echo "Installing flannel ${FLANNEL_VER} ..."
tar xzvf flannel-${FLANNEL_VER}-linux-amd64.tar.gz
chmod +x flanneld
sudo mv flanneld /usr/local/bin/flanneld

cat <<EOF >/tmp/flannel-config.json
{
    "Network": "10.200.0.0/16",
    "SubnetLen": 24,
    "Backend": {
        "Type": "vxlan",
        "Port": 8285
     }
}
EOF

sudo mv /tmp/flannel-config.json /etc/flannel-config.json

# Import default configuration into etcd for maya master
etcdctl --ca-file=/etc/etcd/ca.crt set /coreos.com/network/config < /etc/flannel-config.json

cat <<EOF > /tmp/flanneld.service
[Unit]
Description=Flannel SDN
Documentation=https://github.com/coreos/flannel

[Service]
ExecStart=/usr/local/bin/flanneld \
  --iface=enp0s8 \
  --etcd-cafile=/etc/etcd/ca.crt \
  --etcd-endpoints=https://k8s-master-1:2379
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

#creating flanneld daemon service
sudo mv /tmp/flanneld.service /etc/systemd/system/flanneld.service

# Start flannel
echo "Starting flannel service..."
   sudo systemctl enable flanneld
   sudo systemctl start flanneld

echo "Network configuration verified"
