#!/usr/bin/env bash

set -ex

sudo docker pull openebs/cstor-test:ci
sudo docker run -v /tmp:/tmp openebs/cstor-test:ci

sudo mv /tmp/zrepl /usr/local/bin/zrepl
sudo mv /tmp/zfs /usr/local/bin/zfs
sudo mv /tmp/zpool /usr/local/bin/zpool

sudo mv /tmp/libzfs*.so* /usr/lib
sudo mv /tmp/libnvpair*.so* /usr/lib
sudo mv /tmp/libuutil*.so* /usr/lib
sudo mv /tmp/libzpool*.so* /usr/lib

sudo apt-get install --yes -qq libjemalloc-dev

zrepl start -t 127.0.0.1 > /tmp/tempp.txt &

truncate -s 100MB /tmp/img1.img
truncate -s 100MB /tmp/img2.img
