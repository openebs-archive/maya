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

var (
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
					   "image": "{{.ReplicaImage}}:{{.ImageTag}}"
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
					"image": "{{.ControllerImage}}:{{.ImageTag}}"
				 },
				 {
					"name": "maya-volume-exporter",
					"image": "{{.MExporterImage}}:{{.ImageTag}}"
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
					   "image": "{{.PoolImage}}:{{.ImageTag}}"
					},
					{
					  "name": "cstor-pool-mgmt",
					  "image": "{{.PoolMgmtImage}}:{{.ImageTag}}"
					},
					{
					  "name": "maya-exporter",
					  "image": "{{.MExporterImage}}:{{.ImageTag}}"
					}
				]
			  }
		   }
		}
	  }`

	cstorTargetPatchTemplate = `{
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
					   "name": "cstor-istgt",
					   "image": "{{.IstgtImage}}:{{.ImageTag}}"
					},
					{
					   "name": "maya-volume-exporter",
					   "image": "{{.MExporterImage}}:{{.ImageTag}}"
					},
					{
					   "name": "cstor-volume-mgmt",
					   "image": "{{.VolumeMgmtImage}}:{{.ImageTag}}"
					}
				 ]
			  }
		   }
		}
	 }`
)
