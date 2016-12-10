#!/bin/bash

set -e

echo "Fetching first possible IP address"

ip addr | grep 'state UP' -A2 | tail -n1 | awk '{print $2}' | cut -f1  -d'/'
