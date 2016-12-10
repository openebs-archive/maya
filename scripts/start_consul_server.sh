#!/bin/bash

set -e

echo "Starting consul server"

sudo systemctl enable consul-server.service
sudo systemctl start consul-server.service
