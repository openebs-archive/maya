// Copyright © 2017-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zvol

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/golang/glog"
	col "github.com/openebs/maya/cmd/maya-exporter/app/collector"
	types "github.com/openebs/maya/pkg/exec"
	zvol "github.com/openebs/maya/pkg/zvol/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

// volumeList implements prometheus.Collector interface
type volumeList struct {
	sync.Mutex
	*listMetrics
	request bool
	runner  types.Runner
}

// NewVolumeList returns new instance of volumeList
func NewVolumeList(runner types.Runner) col.Collector {
	return &volumeList{
		listMetrics: newListMetrics().
			withUsedSize().
			withAvailableSize().
			withParseErrorCounter().
			withCommandErrorCounter().
			withRequestRejectCounter().
			withNoDatasetAvailableErrorCounter().
			withInitializeLibuzfsClientErrorCounter(),
		runner: runner,
	}
}

func (v *volumeList) isRequestInProgress() bool {
	return v.request
}

func (v *volumeList) setRequestToFalse() {
	v.Lock()
	v.request = false
	v.Unlock()
}

// collectors returns the list of the collectors
func (v *volumeList) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		v.used,
		v.available,
		v.zfsListParseErrorCounter,
		v.zfsListCommandErrorCounter,
		v.zfsListRequestRejectCounter,
		v.zfsListNoDataSetAvailableErrorCounter,
		v.zfsListInitializeLibuzfsClientErrorCounter,
	}
}

// Describe is implementation of Describe method of prometheus.Collector
// interface.
func (v *volumeList) Describe(ch chan<- *prometheus.Desc) {
	for _, col := range v.collectors() {
		col.Describe(ch)
	}
}

func (v *volumeList) checkError(stdout []byte, ch chan<- prometheus.Metric) error {
	if zvol.IsNotInitialized(string(stdout)) {
		v.zfsListInitializeLibuzfsClientErrorCounter.Inc()
		v.zfsListInitializeLibuzfsClientErrorCounter.Collect(ch)
		return errors.New(zvol.InitializeLibuzfsClientErr.String())
	}

	if zvol.IsNoDataSetAvailable(string(stdout)) {
		v.zfsListNoDataSetAvailableErrorCounter.Inc()
		v.zfsListNoDataSetAvailableErrorCounter.Collect(ch)
		return errors.New(zvol.NoDataSetAvailable.String())
	}
	return nil
}

func (v *volumeList) get(ch chan<- prometheus.Metric) ([]fields, error) {
	v.Lock()
	defer v.Unlock()

	glog.V(2).Info("Run zfs list command")
	stdout, err := zvol.Run(v.runner)
	if err != nil {
		v.zfsListCommandErrorCounter.Inc()
		v.zfsListCommandErrorCounter.Collect(ch)
		return nil, err
	}

	if err := v.checkError(stdout, ch); err != nil {
		return nil, err
	}
	glog.V(2).Infof("Parse stdout of zfs list command, got stdout: \n%v", string(stdout))
	list := listParser(stdout, v.listMetrics)

	return list, nil
}

// Collect is implementation of prometheus's prometheus.Collector interface
func (v *volumeList) Collect(ch chan<- prometheus.Metric) {
	v.Lock()
	if v.isRequestInProgress() {
		v.zfsListRequestRejectCounter.Inc()
		v.Unlock()
		v.zfsListRequestRejectCounter.Collect(ch)
		return
	}

	v.request = true
	v.Unlock()

	volumeLists, err := v.get(ch)
	if err != nil {
		v.setRequestToFalse()
		return
	}

	glog.V(2).Infof("Got zfs list: %#v", volumeLists)
	v.setListStats(volumeLists)
	for _, col := range v.collectors() {
		col.Collect(ch)
	}
	v.setRequestToFalse()
}

func (v *volumeList) setListStats(volumeLists []fields) {
	for _, vol := range volumeLists {
		v.used.WithLabelValues(vol.name).Set(vol.used)
		v.available.WithLabelValues(vol.name).Set(vol.available)
	}
}
