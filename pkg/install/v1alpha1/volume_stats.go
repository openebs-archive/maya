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

package v1alpha1

const volumeStatsYaml = `
---
apiVersion: openebs.io/v1alpha1
kind: CASTemplate
metadata:
  name: cas-volume-stats-default
spec:
  defaultConfig:
  - name: RunNamespace
    value: {{env "OPENEBS_NAMESPACE"}}
  taskNamespace: {{env "OPENEBS_NAMESPACE"}}
  run:
    tasks:
      - cas-volume-stats-default
  output: cas-volume-stats-output-default
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata: 
  name: cas-volume-stats-default
spec:
  meta: |
    id: readvolumesvc
    runNamespace: {{ .Config.RunNamespace.value }}
    apiVersion: v1
    kind: Service
    action: list
    options: |
      labelSelector: openebs.io/persistent-volume={{ .Volume.owner }}
  post: |
      {{- jsonpath .JsonResult "{.items[*].spec.clusterIP}" | trim | saveAs "readvolumesvc.svcIP" .TaskResult | noop -}}
      {{- .TaskResult.readvolumesvc.svcIP | notFoundErr "Volume not found" | saveIf "readstoragepool.notFoundErr" .TaskResult | noop -}}
      {{- $url:= print "http://" .TaskResult.readvolumesvc.svcIP ":9500/metrics/?format=json" -}}
      {{- $store := storeAt .TaskResult -}}
      {{- $runner := storeRunner $store -}}
      {{- get http | withoption "url" $url | withoption "unmarshal" false | runas "getStats" $runner -}}
---
apiVersion: openebs.io/v1alpha1
kind: RunTask
metadata: 
  name: cas-volume-stats-output-default
spec:
  meta: |
    action: output
    id: volumestats
    kind: CASStats
    apiVersion: v1alpha1
  task: |
      {{ .TaskResult.getStats.result | default "" | toString }}
`

// VolumeStatsArtifacts returns the CRDs required for latest version
func VolumeStatsArtifacts() (list artifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamls(VolumeStats{})...)
	return
}

type VolumeStats struct{}

// FetchYamls returns volume stats yamls
func (v VolumeStats) FetchYamls() string {
	return volumeStatsYaml
}
