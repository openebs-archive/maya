#!/bin/bash

set -e

echo "Starting nomad client service ..."

sudo systemctl enable nomad-client.service
sudo systemctl restart nomad-client.service
