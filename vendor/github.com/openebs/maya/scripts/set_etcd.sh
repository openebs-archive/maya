#!/bin/bash

set -e

if [ $# -ne 3 ]; then
    echo usage: $0 SELF_IP SELF_IP_TRIM ETCD_INITIAL_CLUSTER
    exit 1
fi

SELF_IP=$1
SELF_IP_TRIM=$2
ETCD_INITIAL_CLUSTER=$3

echo "Set etcd ..."

# Place systemd service template for etcd
sudo cp /etc/maya.d/templates/etcd.service.tmpl /etc/systemd/system/etcd.service

# Replace the placeholders with actual values
sudo sed -e "s|__SELF_IP_TRIM__|$SELF_IP_TRIM|g" -i /etc/systemd/system/etcd.service
sudo sed -e "s|__SELF_IP__|$SELF_IP|g" -i /etc/systemd/system/etcd.service
sudo sed -e "s|__ETCD_INITIAL_CLUSTER__|$ETCD_INITIAL_CLUSTER|g" -i /etc/systemd/system/etcd.service
