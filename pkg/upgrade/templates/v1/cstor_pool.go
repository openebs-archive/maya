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
	// CSPPatch is used for patching cstor pool
	CSPPatch = `{
    "metadata": {
        "labels": {
            "openebs.io/version": "{{.UpgradeVersion}}"
        }{{if lt .CurrentVersion "1.9.0"}},
        "finalizers": [
            "openebs.io/pool-protection"
        ]{{end}}
    },{{if lt .CurrentVersion "1.8.0"}}
    "spec": {
        "poolSpec": {
            "roThresholdLimit": 85
        }
    },{{end}}
    "versionDetails": {
      "desired": "{{.UpgradeVersion}}",
      "status": {
         "state": "ReconcilePending"
      }
    }
}`
	// CSPDeployPatch is used for patching cstor pool deployment
	CSPDeployPatch = `{
  "metadata": {
    "labels": {
      "openebs.io/version": "{{.UpgradeVersion}}"
    },
    "annotations": {
      "cluster-autoscaler.kubernetes.io/safe-to-evict": "false"
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
                  "timeout 120 zfs set io.openebs:livenesstimestamp=\"$(date +%s)\" cstor-$OPENEBS_IO_CSTOR_ID"
                ]
              },
              "failureThreshold": 3,
              "initialDelaySeconds": 300,
              "periodSeconds": 60,
              "successThreshold": 1,
              "timeoutSeconds": 150
            }{{if lt .CurrentVersion "1.7.0"}},
            "volumeMounts": [
              {
                "name": "storagepath",
                "mountPath": "/var/openebs/cstor-pool"
              },
              {
                "name": "sockfile",
                "mountPath": "/var/tmp/sock"
              }
            ]
          {{end}}
          },
          {
            "name": "cstor-pool-mgmt",
            "image": "{{.PoolMgmtImage}}:{{.ImageTag}}"{{if lt .CurrentVersion "1.7.0"}},
            "volumeMounts": [
              {
                "name": "storagepath",
                "mountPath": "/var/openebs/cstor-pool"
              },
              {
                "name": "sockfile",
                "mountPath": "/var/tmp/sock"
              }
            ]
          {{end}}
          },
          {
            "name": "maya-exporter",
            "image": "{{.MExporterImage}}:{{.ImageTag}}"{{if lt .CurrentVersion "1.7.0"}},
            "volumeMounts": [
              {
                "name": "storagepath",
                "mountPath": "/var/openebs/cstor-pool"
              },
              {
                "name": "sockfile",
                "mountPath": "/var/tmp/sock"
              }
            ]
          {{end}}
	  }
        ]{{if lt .CurrentVersion "1.7.0"}},
        "volumes": [
          {
            "name": "storagepath",
            "hostPath": {
              "path": "{{.BaseDir}}/cstor-pool/{{.SPCName}}",
              "type": "DirectoryOrCreate"
            }
          },
	   {
            "name": "sockfile",
            "emptyDir": {}
	   }
        ]
      {{end}}
      }
    }
  }
}`
)
