#!/bin/bash

set -e

sudo systemctl enable consul-server.service
sudo systemctl start consul-server.service
