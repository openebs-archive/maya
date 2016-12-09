#!/bin/bash

set -e

# Copy the various consul config templates
sudo cp /etc/maya.d/templates/consul-server.json.tmpl /etc/consul.d/server/consul-server.json
sudo cp /etc/maya.d/templates/consul-server-init.conf.tmpl /etc/init/consul-server.conf
