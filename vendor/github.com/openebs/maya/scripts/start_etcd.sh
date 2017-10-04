#!/bin/bash

set -e

echo "Starting etcd service ..."

sudo systemctl daemon-reload
sudo systemctl enable etcd.service
sudo systemctl restart etcd.service
