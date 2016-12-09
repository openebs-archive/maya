#!/bin/bash

set -e

# Copy the consul config template & systemd service template for consul
sudo cp /etc/maya.d/templates/consul-server.json.tmpl /etc/consul.d/server/consul-server.json
sudo cp /etc/maya.d/templates/consul-server.service.tmpl /etc/systemd/system/consul-server.service
