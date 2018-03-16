#!/bin/sh

set -ex

exec /usr/local/bin/cstor-pool-mgmt start --kubeconfig=$HOME/.kube/config
exec service ssh start
exec service rsyslog start
