/*
Copyright 2019 The OpenEBS Authors.

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

import (
	"bytes"

	csp "github.com/openebs/maya/pkg/cstor/pool/v1alpha3"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	pv "github.com/openebs/maya/pkg/kubernetes/persistentvolume/v1alpha1"
	svc "github.com/openebs/maya/pkg/kubernetes/service/v1alpha1"
)

var (
	upgradeVersion = "1.1.0-RC1"
	currentVersion = "1.0.0"

	replicaPatchTemplate = `{
		"metadata": {
		   "labels": {
			  "openebs.io/version": "{{.UpgradeVersion}}",
			  "openebs.io/persistent-volume": "{{.PVName}}",
			  "openebs.io/replica": "jiva-replica"
		   }
		},
		"spec": {
			"selector": {
				"matchLabels":{
					"openebs.io/persistent-volume": "{{.PVName}}",
					"openebs.io/replica": "jiva-replica"
				}
			},
		   "template": {
			   "metadata": {
				   "labels": {
					   "openebs.io/version": "{{.UpgradeVersion}}",
					   "openebs.io/persistent-volume": "{{.PVName}}",
					   "openebs.io/replica": "jiva-replica"
				   }
			   },
			  "spec": {
				 "containers": [
					{
					   "name": "{{.ReplicaContainerName}}",
					   "image": "{{.ReplicaImage}}:{{.UpgradeVersion}}"
					}
				 ],
				 "affinity": {
					 "podAntiAffinity": {
						 "requiredDuringSchedulingIgnoredDuringExecution": [
							 {
								 "labelSelector": {
									 "matchLabels": {
										 "openebs.io/persistent-volume": "{{.PVName}}",
										 "openebs.io/replica": "jiva-replica"
									 }
								 },
					 "topologyKey": "kubernetes.io/hostname"
							 }
						 ]
					 }
				 }
			  }
		   }
		}
	 }`

	targetPatchTemplate = `{
		"metadata": {
		   "labels": {
			 "openebs.io/version": "{{.UpgradeVersion}}"
		   }
		},
		"spec": {
		   "template": {
			  "metadata": {
				 "labels":{
					"openebs.io/version": "{{.UpgradeVersion}}"
				 }
			  },
			 "spec": {
			   "containers": [
				 {
					"name": "{{.ControllerContainerName}}",
					"image": "{{.ControllerImage}}:{{.UpgradeVersion}}"
				 },
				 {
					"name": "maya-volume-exporter",
					"image": "{{.MExporterImage}}:{{.UpgradeVersion}}"
				 }
			   ]
			 }
		   }
		}
	  }`

	openebsVersionPatchTemplate = `{
		"metadata": {
		   "labels": {
			  "openebs.io/version": "{{.}}"
		   }
		}
	 }`

	cspDeployPatchTemplate = `{
		"metadata": {
		   "labels": {
			  "openebs.io/version": "{{.UpgradeVersion}}"
		   }
		},
		"spec": {
		   "template": {
			  "metadata": {
				 "labels": {
					"openebs.io/version": "{{.UpgradeVersion}}"
				 }
			  },
			  "spec": {
				 "containers": [
					{
					   "name": "cstor-pool",
					   "image": "{{.PoolImage}}:{{.UpgradeVersion}}"
					},
					{
					  "name": "cstor-pool-mgmt",
					  "image": "{{.PoolMgmtImage}}:{{.UpgradeVersion}}"
					},
					{
					  "name": "maya-exporter",
					  "image": "{{.MExporterImage}}:{{.UpgradeVersion}}"
					}
				]
			  }
		   }
		}
	  }`

	buffer bytes.Buffer

	deployClient  = deploy.NewKubeClient()
	serviceClient = svc.NewKubeClient()
	pvClient      = pv.NewKubeClient()
	cspClient     = csp.KubeClient()
)

// Exec ...
func Exec(kind, name, openebsNamespace string) error {
	// TODO
	// verify openebs namespace and check maya-apiserver version
	switch kind {
	case "jivaVolume":
		err := jivaUpgrade(name, openebsNamespace)
		if err != nil {
			return err
		}
	case "storagePoolClaim":
		err := spcUpgrade(name, openebsNamespace)
		if err != nil {
			return err
		}
	case "cstorPool":
		err := cspUpgrade(name, openebsNamespace)
		if err != nil {
			return err
		}
	default:
		return errors.Errorf("Invalid kind for upgrade")
	}
	return nil
}
