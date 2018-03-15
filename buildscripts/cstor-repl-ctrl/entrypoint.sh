#!/bin/sh

set -ex

exec /usr/local/bin/cstor-repl-ctrl start --kubeconfig=$HOME/.kube/config
exec service ssh start
exec service rsyslog start
