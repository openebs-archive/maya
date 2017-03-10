#!/bin/bash

set -e

SELF_IPV4=$1

if [ $# -eq 0 ]; then

   SELF_IP4=127.0.0.1
fi
echo "Setting up Mayaserver Daemon ...with ip $SELF_IPV4"

# Place systemd service template for Mayaserver
sudo cp /etc/maya.d/templates/mayaserver.service.tmpl /etc/systemd/system/mayaserver.service
sudo cp /etc/maya.d/templates/nomad_global.INI.tmpl /etc/mayaserver/orchprovider/nomad_global.INI

# Replace the placeholders with actual values
sudo sed -e "s|__SELF_IPV4__|$SELF_IPV4|g" -i /etc/systemd/system/mayaserver.service 

sudo sed -e "s|__SELF_IPV4__|$SELF_IPV4|g" -i /etc/mayaserver/orchprovider/nomad_global.INI

echo "Starting Mayaserver service ..."

sudo systemctl enable mayaserver.service
sudo systemctl restart mayaserver.service
