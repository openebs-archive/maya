/*
Copyright 2018 The OpenEBS Authors

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

// TODO
// Rename this file by removing the version suffix information
package v1alpha1

const openEBSCRDYamls = `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: castemplates.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Cluster
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: castemplates
    # singular name to be used as an alias on the CLI and for display
    singular: castemplate
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: CASTemplate
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - cast
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: runtasks.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Namespaced
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: runtasks
    # singular name to be used as an alias on the CLI and for display
    singular: runtask
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: RunTask
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - rtask
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: storagepoolclaims.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Cluster
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: storagepoolclaims
    # singular name to be used as an alias on the CLI and for display
    singular: storagepoolclaim
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: StoragePoolClaim
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - spc
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: storagepools.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Cluster
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: storagepools
    # singular name to be used as an alias on the CLI and for display
    singular: storagepool
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: StoragePool
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - sp
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: cstorpools.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Cluster
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: cstorpools
    # singular name to be used as an alias on the CLI and for display
    singular: cstorpool
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: CStorPool
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - csp
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: cstorvolumes.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Namespaced
  names:
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: CStorVolume
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: cstorvolumes
    # singular name to be used as an alias on the CLI and for display
    singular: cstorvolume
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - cstorvolume
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: cstorvolumereplicas.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Namespaced
  names:
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: CStorVolumeReplica
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: cstorvolumereplicas
    # singular name to be used as an alias on the CLI and for display
    singular: cstorvolumereplica
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - cvr
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: disks.openebs.io
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: openebs.io
  # version name to use for REST API: /apis/<group>/<version>
  version: v1alpha1
  # either Namespaced or Cluster
  scope: Cluster
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: disks
    # singular name to be used as an alias on the CLI and for display
    singular: disk
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: Disk
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - disk
---
`

// OpenEBSCRDArtifacts returns the CRDs required for latest version
func OpenEBSCRDArtifacts() (list artifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamls(openEBSCRDs{})...)
	return
}

type openEBSCRDs struct{}

// FetchYamls returns all the CRD yamls related to 0.7.0
// in a string format
//
// NOTE:
//  This is an implementation of MultiYamlFetcher
func (o openEBSCRDs) FetchYamls() string {
	return openEBSCRDYamls
}
