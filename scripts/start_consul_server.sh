#!/bin/bash

set -e

echo "Starting consul server service ..."

sudo systemctl enable consul-server.service
sudo systemctl restart consul-server.service
