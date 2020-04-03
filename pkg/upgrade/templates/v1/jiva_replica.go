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

package templates

var (
	// JivaReplicaPatch is generic template for replica patch
	JivaReplicaPatch = `{
		"metadata": {
		   "labels": {
			  "openebs.io/version": "{{.UpgradeVersion}}",
			  "openebs.io/persistent-volume": "{{.PVName}}",
			  "openebs.io/replica": "jiva-replica"
		   }
		},
		"spec": {
			"replicas": 1,
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
)
