### Problems solved
* OpenEBS requires open-iscsi, fsck and other filesystem related tools to be installed on nodes. These dependencies are currently being installed manually.OS dependency (This needs to be looked into as to how this can be resolved)
* Read-Only issue: A health check thread will be continuously running in CSI node which will monitor the mount points. As soon as a volume is published(mounted) at a node, the mount point is sent to health check thread for monitoring. If the thread finds that the mount point is in RO mode, it does an unmount and mount operation.
* Multi-attach problem: The CSI node which gets a request to publish the volume performs the following steps:
    - Confirms at other CSI nodes that the volume has been unpublished(unmounted)
        - (TODO) Steps taken if the other CSI nodes are not able to unpublish
    - Registers itself at istgt(use of initiator groups) to accept connections from this initiator and the other initiator is automatically unregistered

### Storage vendor CSI driver components
* CSI Controller plugin(StatefulSet)
* CSI Node plugin(Daemon set in privileged mode -- Required because of mount operation)

### Pods that will be scheduled
* openebs-csi-controller-driver (CSI Controller plugin)
    - Sidecars:
        - liveness probe
        - external-provisioner
            - Watches the Kubernetes API server for PersistentVolumeClaim objects.
            - Calls CreateVolume CSI driver API
            - Once a new volume is successfully provisioned, the sidecar container creates a Kubernetes PersistentVolume object to represent the volume.
            - It also supports the Snapshot DataSource. If a Snapshot CRD is specified as a data source on a PVC object, the sidecar container fetches the information about the snapshot by fetching the SnapshotContent object and populates the data source field in the resulting CreateVolume call to indicate to the storage system that the new volume should be populated using the specified snapshot.
        - cluster driver registrar
          - Registers csi driver with a kubernetes cluster by creating a csi driver object
          - Args used: pod-info-mount-version, driver-requires-attachment
        - external-snapshotter
            - To support provisioning volume snapshots and the ability to provision new volumes using those snapshots.
            - Watches the Kubernetes API server for VolumeSnapshot and VolumeSnapshotContent CRD objects
        - external-resizer
* openebs-csi-node-driver (CSI Node plugin)
    - Sidecars:
        - node-driver-registrar
            - Fetches driver information (using NodeGetInfo)
            - Kubelet directly issues CSI NodeGetInfo, NodeStageVolume, and NodePublishVolume calls against CSI drivers.
        - external-attacher
            - Not required in Kubernetes CSI GA version 
        - liveness probe
            - Optional

##### CSI Driver Object
This object allows CSI drivers to specify how Kubernetes should interact with it.

```
apiVersion: csi.storage.k8s.io/v1alpha1
kind: CSIDriver
metadata:
  name: openebs-csi.openebs.io
spec:
  attachRequired: false
  podInfoOnMountVersion: ""

```
* Name: Name of the CSI Driver
* attachRequired: Required when ControllerPublishVolume method is implemented (Default is true)
* podInfoOnMountVersion: Indicates this CSI volume driver requires additional pod information (like pod name, pod UID, etc.)  during mount operations. No info is sent by default. If value is set to a valid version, Kubelet will pass pod information as volume_context in CSI NodePublishVolume calls.(https://kubernetes-csi.github.io/docs/csi-driver-object.html)

#### CSI Node Object
The kubelet will automatically populate the CSINodeInfo object for the CSI driver as part of kubelet plugin registration
CSI drivers generate node specific information. Instead of storing this in the Kubernetes Node API Object.
  - Maps Kubernetes node name to CSI Node name
  - A way for kubelet to communicate to the kube-controller-manager and kubernetes scheduler whether the driver is available (registered) on the node or not.
  - Volume topology
```
apiVersion: csi.storage.k8s.io/v1alpha1
kind: CSINodeInfo
metadata:
  name: node1
spec:
  drivers:
  - name: openebs-csi.openebs.io
    available: true
    volumePluginMechanism: csi-plugin
status:
  drivers:
  - name: openebs-csi.openebs.io
    nodeID: storageNodeID1
    topologyKeys: []
```


### Interfaces to be implemented
#### CSI Identity
* GetPluginInfo
* GetPluginCapabilities
* Probe

#### CSI Controller
* Provision(Create)/Delete Volume (REST call to m-apiserver /latest/volumes/)
* Publish/Unpublish Volume on some required node(In case of OpenEBS, this boils down to iscsi login/logout on a particular node)-This will return success without doing anything and login/logout will be done in the nodeStageVolume/nodeUnStageVolume itself). Although this can be used to avoid multiattach issues which will be discussed later.
* Validate volume capabilities
* List Volumes(REST call to m-apiserver /latest/volumes/)
* Get controller capabilities
* Get Capacity of the total available storage pool
* Create/List/Delete Snapshot (REST call to m-apiserver /latest/snapshots/)

#### CSI Node
* NodeStageVolume / NodeUnstageVolume (login, format, temporarily mount/unmount the volume to a staging path)
* NodePublishVolume / NodeUnpublishVolume (mount/unmount volume from staging to target path -- Bind mount)
* NodeGetVolumeStats
* NodeGetId (Return unique ID of the node)
* NodeGetCapabilities
* NodeGetInfo

All the above 3 components will be put in the same binary
    
### Usage
[csi-driver-usage](https://github.com/payes/maya/blob/csi-driver-design/docs/csi-driver-Usage.md)

### FAQs
* How does CSI know that target/pool pods are healthy to serve IOs or it just keeps retying re-mounts?
    - Unless the publish(mount) is not successful, kubernetes will not start the application(need to verify this). It keeps on retrying login and mount unless the volume is healthy. One more interesting experiment that can be explored is that we can pass on vedor related iscsi opcodes to the device to get info directly from istgt.
* How does the user know that the volume went to RO state, and CSI is working on it to re-mount ?
    - This question has not been thought of till now, this can be discussed after some basic implementation is done.
* Can the above info be put in a CR ?
    - This can be worked out.
* Can CSI driver be asked to unmount a volume?
    - When the kubernetes shifts an app from one node to other, it asks the CSI driver to unmount and remount at the other node. Any of OpenEBS components should not need to trigger this.
* Can CSI driver be asked to stop remounting a volume ?
    - This control should not be touched by OpenEBS. Although a REST API can also be exposed to perform this.

### References
* https://medium.com/google-cloud/understanding-the-container-storage-interface-csi-ddbeb966a3b
* https://arslan.io/2018/06/21/how-to-write-a-container-storage-interface-csi-plugin/
* https://events.linuxfoundation.org/wp-content/uploads/2017/12/Internals-of-Docking-Storage-with-Kubernetes-Workloads-Dennis-Chen-Arm.pdf
* https://github.com/hetznercloud/csi-driver
* https://kubernetes-csi.github.io/docs/print.html
* https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/container-storage-interface.md#recommended-mechanism-for-deploying-csi-drivers-on-kubernetes

