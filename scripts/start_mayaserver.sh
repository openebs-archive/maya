#!/bin/bash

set -e

echo "Setup Mayaserver Daemon ..."

# Place systemd service template for Mayaserver
sudo cp /etc/maya.d/templates/mayaserver.service.tmpl /etc/systemd/system/mayaserver.service

echo "Starting Mayaserver service ..."

sudo systemctl enable mayaserver.service
sudo systemctl restart mayaserver.service
