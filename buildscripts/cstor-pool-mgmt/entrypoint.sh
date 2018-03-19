#!/bin/sh

set -ex

rm /usr/local/bin/zrepl
exec /usr/local/bin/cstor-pool-mgmt start
exec service ssh start
exec service rsyslog start
