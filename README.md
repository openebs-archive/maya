## Maya

Maya is the storage orchestration system for managing storage for millions of containers. Maya can accomplish complex storage management tasks with deceptive simplicity. Maya can manage storage across multiple Maya-Realms (aka clusters/environments), that are co-located or geographically seperated and can also run from within a single host. 

TODO: Link to the Architecture Diagram. 

Any \*inx host that supports container can be converted into Maya-Storage-Host using the following commands. 

## Usage

### Initialize 

Initialize the default configuration file for adding the Host to new Maya Realm.
```
maya init
```

### Load Maya
Start the maya services based on the configuration specified in the /etc/maya.conf. 
```
maya load
```
Default maya.conf is created by **_maya init_** and can be modified to perform any of the following:
- Run as Storage Host only in a new or existing Maya Realm
- Run Maya orchestration services only
- Run both Maya orchestration services and also configure as Storage Host in a new or existing Maya Realm

**_maya load_** can also be used to re-load the configuration from maya.conf. 

### Use Maya for Managing Volumes
The following command will create a Maya volume using the default *volume spec*. 
```
maya volume <name> [--spec <volume-spec>]
```
Each volume in maya is associated with a spec that defines the features for the volume like the capacity, *jiva* version, persistent store, etc., The default volume spec will create a new container and expose an volume with :
- capacity 100GB
- latest Jiva version 
- data persisted to the directory /opt/maya-store/<vol-name>. 

The volume will be accessible via iSCSI and can be connected from local host. 

**/opt/maya-store** is the default store created when Maya is installed on the local disk. This can ge changed via the /etc/maya.conf to specify a different directory or disks (in case of single node setups) or can be an shared storage (in case of clustered setups).

## Build and Install Maya

TBD

## Forums

Join us at our [gitter](https://gitter.im/openebs/Lobby) lobby.
