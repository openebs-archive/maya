## Maya

Maya is the storage orchestration system for managing persistent storage for millions of containers. Maya can accomplish complex storage management tasks with deceptive simplicity. 

TODO: Link to the Architecture Diagram. 

Maya can manage storage across multiple Maya-Realms (or clusters or environements), that are co-located or geographically seperated. 

## Usage

### Initialize 

Initialize the default configuration file for adding the Host to new Maya Realm.
```
maya init
```

### Load Maya
Start the maya services based on the configuration specified in the /etc/maya.conf. The maya.conf is created by maya init and can be modified to specify if this Host needs to be run in its own Realm or attached it to one of the existing realms. 
```
maya load
```

## Build and Install Maya

