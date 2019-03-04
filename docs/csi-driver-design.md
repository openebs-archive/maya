### Problems solved
* OpenEBS requires open-iscsi, fsck and other filesystem related tools to be installed on nodes. These dependencies are currently being installed manually.OS dependency
* Read-Only issue: We will keep a health check thread continuously running in CSI node which will monitor the mount point. As soon as a volume is published(mounted) at a node, the mount point is sent to health check thread for monitoring. If the thread finds that the mount point is in RO mode, it does a remount operation.
* Multi-attach problem: The CSI node which gets a request to publish the volume performs the following steps:
    - Confirms at other CSI nodes that the volume has been unpublished(unmounted)
        - (TODO) Steps taken if the other CSI nodes are not able to unpublish
    - Registers itself at istgt(use of initiator groups) to accept connections from this initiator and the other initiator is automatically unregistered

### Storage vendor CSI driver components
* CSI Controller plugin(StatefulSet)
* CSI Node plugin(Daemon set in privileged mode -- Required because of mount operation)

### Interfaces
#### CSI Identity
* GetPluginInfo
* GetPluginCapabilities
* Probe

#### CSI Controller
* Provision(Create)/Delete Volume (REST call to m-apiserver /latest/volumes/)
* Publish/Unpublish Volume on some required node(In our case this boils down to iscsi login/logout on a particular node)-This will return success without doing anything and we will login/logout in the nodeStageVolume itself). Although this can be used to avoid multiattach issues which will be discussed later.
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

### Implementation Details

### Pre-requisites
* Make sure:
    - Following feature gates are enabled in your cluster (kubelet and kube-apiserver):
```--feature-gates=CSINodeInfo=true,CSIDriverRegistry=true```
    - Privileged pods are allowed in your cluster (kubelet and kube-apiserver:
```--allow-privileged=true```
* Create the CSINodeInfo resource:
    `kubectl create -f https://raw.githubusercontent.com/kubernetes/csi-api/master/pkg/crd/manifests/csinodeinfo.yaml`
    
### Usage
[csi-driver-usage](https://github.com/payes/maya/blob/csi-driver-design/docs/csi-driver-Usage.md)

### FAQs
* How does CSI knows that target/pool pods are healthy to server IOs or it just keeps retying re-mounts?
    - Unless the publish(mount) is not successful, kubernetes will not start the application(need to verify this). It keeps on retrying login and mount unless the volume is healthy. One more interesting experiment that can be explored is that we can pass on vedor related iscsi opcodes to the device to get info directly from istgt
* How user knows that volume went to RO state, and CSI is working on it to re-mount it?
    - This question has not been thought of till now, we can discuss this after some basic implementation is done.
* Can the above info be put in a CR ?
    - This can be worked out.
* Can we ask CSI to unmount a volume?
    - When the kubernetes shifts an app from one node to other, it asks the CSI driver to unmount and remount at the other node. I think any of openebs component should not need to trigger this.
* Can we ask CSI to stop remounting a volume ?
    - This control should not be touched by us. Although a REST API can also be exposed to perform this.

### References
* https://medium.com/google-cloud/understanding-the-container-storage-interface-csi-ddbeb966a3b
* https://arslan.io/2018/06/21/how-to-write-a-container-storage-interface-csi-plugin/
* https://events.linuxfoundation.org/wp-content/uploads/2017/12/Internals-of-Docking-Storage-with-Kubernetes-Workloads-Dennis-Chen-Arm.pdf
* https://github.com/hetznercloud/csi-driver
* https://kubernetes-csi.github.io/docs/print.html
* https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/container-storage-interface.md#recommended-mechanism-for-deploying-csi-drivers-on-kubernetes

