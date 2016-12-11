#!/bin/bash

set -e

echo "Starting nomad server service ..."

sudo systemctl enable nomad-server.service
sudo systemctl start nomad-server.service
