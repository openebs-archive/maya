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
	// CSPDeployPatch is used for patching cstor pool deployment
	CSPDeployPatch = `{
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
					   "image": "{{.PoolImage}}:{{.ImageTag}}",
					   "livenessProbe": {
                            "exec": {
                                "command": [
                                    "/bin/sh",
                                    "-c",
                                    "zfs set io.openebs:livenesstimestamp=\"$(date +%s)\" cstor-$OPENEBS_IO_CSTOR_ID"
                                ]
                            }
                        }
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
)
