#!/bin/bash -x

ETCD_VER=v3.0.13
DOWNLOAD_URL=https://github.com/coreos/etcd/releases/download




if [ ! -f /usr/local/bin/etcd ]
then
  curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o ~/etcd-${ETCD_VER}-linux-amd64.tar.gz
    tar xzvf ~/etcd-${ETCD_VER}-linux-amd64.tar.gz -C /usr/local/bin --strip-components=1
  chmod +x /usr/local/bin/etcd
fi

IP=`hostname -I | awk '{ print $2 }'`

ETCD0=192.168.60.10
ETCD1=192.168.60.11
ETCD2=192.168.60.12

case $IP in
$ETCD0) NAME=etcd0;;
$ECTD1) NAME=etcd1;;
$ECTD2) NAME=etcd2;;
esac

etcd --name $NAME \
  --initial-advertise-peer-urls=http://$IP:2380 \
  --listen-peer-urls=http://$IP:2380 \
  --listen-client-urls=http://$IP:2379,http://127.0.0.1:2379 \
  --advertise-client-urls=http://$IP:2379 \
  --initial-cluster-token=etcd-cluster-1 \
  --initial-cluster=etcd0=http://$ETCD0:2380,etcd1=http://$ETCD1:2380,etcd2=http://$ETCD2:2380 \
  --initial-cluster-state=new
