## Maya

[![Build Status](https://travis-ci.org/openebs/maya.svg?branch=master)](https://travis-ci.org/openebs/maya)


Maya is the storage orchestration system for managing storage for millions of containers. 
Maya can accomplish complex storage management tasks with deceptive simplicity. Maya can 
manage storage across multiple Maya-Realms (aka clusters/environments), that are co-located
or geographically seperated and can also run from within a single host. 

Well, the *story* does not end here. Maya manages itself i.e. `maya manages maya`. 
It has been a long desire to make the lives of storage operators easier. Maya has been 
designed ground up to implement just this i.e. manage & tune its own devops requirements. 

![Quick-glance overview](https://github.com/openebs/openebs/blob/master/docs/MayaArchitectureOverview.png) 

Any \*inx host that supports container can be converted into OpenEBS Storage Host using maya. 


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


## Setup and Initialize

NOTE: When there are multiple IPs on the machine, you can specify the IP to be used for management traffic using **-self-ip**

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


### Load Maya
Start the maya services based on the configuration specified in the /etc/maya.conf. 
```
maya load
```
Default maya.conf is created by **_maya init_** and can be modified to perform any of
the following:
- Run as Storage Host only in a new or existing Maya Realm
- Run Maya orchestration services only
- Run both Maya orchestration services and also configure as Storage Host in a new or 
existing Maya Realm

**_maya load_** can also be used to re-load the configuration from maya.conf. 

### Use Maya for Managing Volumes
The following command will create a Maya volume using the default *volume spec*. 
```
maya volume <name> [--spec <volume-spec>]
```
Each volume in maya is associated with a spec that defines the features for the volume 
like the capacity, *jiva* version, persistent store, etc., The default volume spec will
create a new container and expose an volume with :
- capacity 100GB
- latest Jiva version 
- data persisted to the directory /opt/maya-store/<vol-name>. 

The volume will be accessible via iSCSI and can be connected from local host. 

**/opt/maya-store** is the default store created when Maya is installed on the local disk. 
This can ge changed via the /etc/maya.conf to specify a different directory or disks 
(in case of single node setups) or can be an shared storage (in case of clustered setups).


## Forums

Join us at our [gitter](https://gitter.im/openebs/Lobby) lobby.
