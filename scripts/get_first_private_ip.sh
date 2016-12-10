#!/bin/bash

set -e

# NOTE - Do not echo any thing else in this script
# The output of below command is used as-is

ip addr | grep 'state UP' -A2 | tail -n1 | awk '{print $2}' | cut -f1  -d'/'
