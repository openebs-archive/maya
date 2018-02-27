#!/bin/sh

set -ex

#exec cat /usr/local/bin/tgt
#/usr/local/bin/tgt & 

sudo pkill tgt
service ssh start
service rsyslog start
cstor-sidecar start
child=$!

wait
#exec for ((i = 0; ; i++)); do echo "$i: $(date)"; sleep 100; done;
#exec /bin/bash -c "trap : TERM INT; sleep infinity & wait"

