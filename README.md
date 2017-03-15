## Overview

[![Build Status](https://travis-ci.org/openebs/maya.svg?branch=master)](https://travis-ci.org/openebs/maya)

OpenEBS Storage Orchestration, abstracts the operational burden of setting up and managing the Storage (Persistent Volumes) and Storage Infrastructure (Physical Storage - Local/Cloud Disks/SSDs/Cache, Controllers, Networks). OpenEBS Orchestration is delivered through a set of services and tools that seamlessly integrate OpenEBS into your container eco-system. 

OpenEBS Storage Orchestration allows you to manage storage across multiple zones (aka clusters/environments), that are co-located or geographically seperated and can also run from within a single host. Maya can move the storage across different tiers based on the application needs (volume migration). OpenEBS learns and adapts to the changing environement through machine learning and data analytics. 

OpenEBS aims at managing storage for millions of containers with deceptive simplicity. 

Maya (meaning magic) is the command line interface for setting up and managing the OpenEBS Storage Orchestration services.

## Design

Maya helps in managing the OpenEBS Storage Orchestration services, that can be classified based on where they are deployed as follows:

![Maya Design](https://github.com/openebs/openebs/blob/master/documentation/source/_static/maya-hld.png)

- Control Plane Components - like API Server (mAPI) that helps in processing the requests for creating new volumes. The API server, will use the container orchestration engines to deploy the VSM. Another example is the Analytics (mAnalytics) engine that gathers the data via the machine learning probes and runs heuristics analysis to optimize storage deployment. In hyper-converged deployment modes (example with Kubernetes), these services will be deployed along-side kubernetes master nodes. When deployed with Kubernetes (mSCH) is kubernetes-scheduler that is plugged with the storage metrics for placement of VSMs. 

- Node/Host Components - are the services and tools that run on the Container hosts where the VSM containers are scheduled. Services (like mStorageInterace) on the node help in managing the disks attached to the hosts and help in carving out the required disks to different VSMs. There is also an agent that runs on the nodes for interfacing with the control plane, or to external providers like terraform or amazon webservices. 

Maya is primarily designed to : 
- Install and configure Services based on node type, like OpenEBS Master, OpenEBS Storage Host, K8s-master, K8s-minion, etc., 
- Install and configure OpenEBS storage orchestration services like mayaserver, mAnalytics, etc,. 
- Install and configure network infrastructure services/plugins like flannel, weave, etc,. 
- Install and configure Kubernetes services like kube-apiserver, kube-proxy, kube-scheduler, etc., 
- Create and Manage OpenEBS VSMs

Maya aims at making the infrastructure programmable via the yaml files. Once Maya is installed, it can read the node infrastructure intent (speficied in yaml file) and will install the required components. 

## Install

## Installing Maya from binaries

Pre-requisites : ubuntu 16.04, wget, unzip

```
RELEASE_TAG=0.0.3
wget https://github.com/openebs/maya/releases/download/${RELEASE_TAG}/maya-linux_amd64.zip
unzip maya-linux_amd64.zip
sudo mv maya /usr/bin
rm -rf maya-linux_amd64.zip
```

## Installing Maya from source

Pre-requisites : ubuntu 16.04, git, zip, unzip, go. 

```
mkdir -p $GOPATH/src/github.com/openebs && cd $GOPATH/src/github.com/openebs
git clone https://github.com/openebs/maya.git
cd maya && make dev
```


## Usage

#### Setup OpenEBS Maya Master (omm)

Example : Assuming 172.28.128.3 is where you require the management server to be running. 
```
ubuntu@master-01:~$ maya setup-omm -self-ip=172.28.128.3
```

#### Setup OpenEBS Storage Host (osh)

Example : Assuming Maya Master is reachable on 172.28.128.3 and you would like the OpenEBS Storage Host to communicate using 172.28.128.6
```
ubuntu@host-01:~$ maya setup-osh -self-ip=172.28.128.6 -omm-ips=172.28.128.3
```

