/*
Copyright 2019 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

const localPVSCYamls = `
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-hostpath
  annotations:
    #Define a new CAS Type called "local"
    #which indicates that Data is stored 
    #directly onto hostpath. The hostpath can be:
    #- device (as block or mounted path)
    #- hostpath (sub directory on OS or mounted path)
    openebs.io/cas-type: local
    cas.openebs.io/config: |
      - name: StorageType
        value: "hostpath"
      - name: BasePath
        value: "/var/openebs/local"
provisioner: openebs.io/local
volumeBindingMode: WaitForFirstConsumer
reclaimPolicy: Delete
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-device
  annotations:
    #Define a new CAS Type called "local"
    #which indicates that Data is stored 
    #directly onto hostpath. The hostpath can be:
    #- device (as block or mounted path)
    #- hostpath (sub directory on OS or mounted path)
    openebs.io/cas-type: local
    cas.openebs.io/config: |
      - name: StorageType
        value: "device"
provisioner: openebs.io/local
volumeBindingMode: WaitForFirstConsumer
reclaimPolicy: Delete
---
`

// LocalPVArtifacts returns the default Local PV storage
// class related artifacts corresponding to latest version
func LocalPVArtifacts() (list artifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamls(localPVSCs{})...)
	return
}

type localPVSCs struct{}

// FetchYamls returns all the yamls related to local pv storage classes
// in a string format
//
// NOTE:
//  This is an implementation of MultiYamlFetcher
func (j localPVSCs) FetchYamls() string {
	return localPVSCYamls
}
