#!/bin/sh

set -ex

/usr/local/bin/cstor-volume-mgmt start &
exec /usr/local/bin/cstor-volume-grpc server start
