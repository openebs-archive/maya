#!/bin/bash

set -e

echo "Starting consul client service ..."

sudo systemctl enable consul-client.service
sudo systemctl restart consul-client.service
