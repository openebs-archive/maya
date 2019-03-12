package zvol

import (
	"sync"
	"time"

	"github.com/golang/glog"
	zvol "github.com/openebs/maya/pkg/zvol/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

// volumeList implements prometheus.Collector interface
type volumeList struct {
	sync.Mutex
	listMetrics
	request bool
}

// NewVolumeList returns new instance of volumeList
func NewVolumeList() *volumeList {
	return &volumeList{
		listMetrics: newListMetrics(),
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
	}
}

// Describe is implementation of Describe method of prometheus.Collector
// interface.
func (v *volumeList) Describe(ch chan<- *prometheus.Desc) {
	for _, col := range v.collectors() {
		col.Describe(ch)
	}
}

func (v *volumeList) get(ch chan<- prometheus.Metric) ([]fields, error) {
	v.Lock()
	defer v.Unlock()
	var timeout = 30 * time.Second

	glog.V(2).Info("Run zfs list command")
	stdout, err := zvol.Run(timeout, runner, "list", "-Hp")
	if err != nil {
		v.zfsListCommandErrorCounter.Inc()
		v.zfsListCommandErrorCounter.Collect(ch)
		return nil, err
	}

	glog.V(2).Infof("Parse stdout of zfs list command, got stdout: \n%v", string(stdout))
	list := listParser(stdout, &v.listMetrics)

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
