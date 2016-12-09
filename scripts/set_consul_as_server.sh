#!/bin/bash

set -ex

# Copy the various consul config templates
cp /etc/maya.d/templates/consul-server.json.tmpl /etc/consul.d/server/consul-server.json
cp /etc/maya.d/templates/consul-server-init.conf.tmpl /etc/init/consul-server.conf
