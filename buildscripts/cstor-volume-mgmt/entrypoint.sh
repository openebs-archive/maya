#!/bin/sh

set -ex

/usr/local/bin/cstor-volume-mgmt start &
service ssh start &
service rsyslog start &
exec /usr/local/bin/cstor-volume-grpc server start
