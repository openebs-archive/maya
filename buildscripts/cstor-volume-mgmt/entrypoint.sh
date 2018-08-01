#!/bin/sh

set -ex

exec /usr/local/bin/cstor-volume-mgmt start
exec service ssh start
exec service rsyslog start
