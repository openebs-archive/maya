### Usage
* Enable following feature gates and parameters in the cluster both in kubelet and kube-apiserver:
    - CSINodeInfo=true
    - CSIDriverRegistry=true
    - KubeletPluginsWatcher=true
    - allow-privileged=true
    - To enable the above feature gates in Kubelet add the foloowing line to /etc/default/kubelet in all the nodes and restart kubelet
        - KUBELET_EXTRA_ARGS=--feature-gates=CSINodeInfo=true,CSIDriverRegistry=true,KubeletPluginsWatcher=true --allow-privileged=true
    - To enable the above feature gates in Kube-apiserver add the foloowing lines to ./kubernetes/manifests/kube-apiserver.yaml in the master node and restart kube-apiserver. These lines have to be added in spec->containers->command: section
        - --feature-gates=CSINodeInfo=true,CSIDriverRegistry=true,KubeletPluginsWatcher=true
        - --allow-privileged=true
* Apply csidriver and csinodeinfo crds
    - kubectl apply -f https://raw.githubusercontent.com/kubernetes/kubernetes/master/cluster/addons/storage-crds/csidriver.yaml 
    - kubectl apply -f https://raw.githubusercontent.com/kubernetes/kubernetes/master/cluster/addons/storage-crds/csinodeinfo.yaml
* Apply openebs-csi-operator yaml
* Create Storage class pointing to openebs csi-driver
    - provisioner: csi-driver.example.com
* Create PVC with the above Storage Class
