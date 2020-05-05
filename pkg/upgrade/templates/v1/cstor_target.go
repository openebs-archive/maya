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
	// CstorTargetPatch is used to patch target deployment for cstor volume
	CstorTargetPatch = `{
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
            "image": "{{.IstgtImage}}:{{.ImageTag}}"{{if isCurrentLessThanNewVersion .CurrentVersion "1.7.0"}},
            "volumeMounts": [
              {
                "name": "storagepath",
                "mountPath": "/var/openebs/cstor-target"
              }
            ]
          {{end}}
          },
          {
            "name": "maya-volume-exporter",
            "image": "{{.MExporterImage}}:{{.ImageTag}}"{{if isCurrentLessThanNewVersion .CurrentVersion "1.7.0"}},
            "volumeMounts": [
              {
                "name": "storagepath",
                "mountPath": "/var/openebs/cstor-target"
              }
            ]
          {{end}}
          },
          {
            "name": "cstor-volume-mgmt",
            "image": "{{.VolumeMgmtImage}}:{{.ImageTag}}"{{if isCurrentLessThanNewVersion .CurrentVersion "1.7.0"}},
            "volumeMounts": [
              {
                "name": "storagepath",
                "mountPath": "/var/openebs/cstor-target"
              }
            ]
          {{end}}
          }
        ]{{if isCurrentLessThanNewVersion .CurrentVersion "1.7.0"}},
        "volumes": [
          {
            "name": "storagepath",
            "hostPath": {
              "path": "{{.BaseDir}}/cstor-target/{{.PVName}}",
              "type": "DirectoryOrCreate"
            }
          }
        ]
      {{end}}
      }
    }
  }
}`
)
