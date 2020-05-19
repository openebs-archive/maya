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

package collector

import (
	"sync"

	v1 "github.com/openebs/maya/pkg/stats/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog"
)

// collector implements prometheus.Collector interface
type collector struct {
	sync.Mutex
	request bool
	Volume
	metrics
}

func (c *collector) isRequestInProgress() bool {
	return c.request
}

func (c *collector) setRequestToFalse() {
	c.Lock()
	c.request = false
	c.Unlock()
}

func New(vol Volume) *collector {
	typ := casType(vol)
	if typ == "" {
		klog.Fatal("exiting...")
	}
	return &collector{
		Volume:  vol,
		metrics: Metrics(typ),
	}
}

func casType(vol Volume) string {
	switch typ := vol.(type) {
	case *jiva:
		return "jiva"
	case *cstor:
		return "cstor"
	default:
		klog.Error("Unknown cas type: ", typ)
		return ""
	}
}

// collectors returns the list of the collectors
func (c *collector) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		c.reads,
		c.writes,
		c.totalReadBytes,
		c.totalWriteBytes,
		c.totalReadTime,
		c.totalWriteTime,
		c.totalReadBlockCount,
		c.totalWriteBlockCount,
		c.actualUsed,
		c.logicalSize,
		c.sectorSize,
		c.sizeOfVolume,
		c.volumeStatus,
		c.connectionErrorCounter,
		c.connectionRetryCounter,
		c.parseErrorCounter,
		c.totalReplicaCounter,
		c.degradedReplicaCounter,
		c.healthyReplicaCounter,
		c.volumeUpTime,
		c.isClientConnected,
		c.targetRejectRequestCounter,
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent. The sent descriptors fulfill the
// consistency and uniqueness requirements described in the Desc
// documentation. (It is valid if one and the same Collector sends
// duplicate descriptors. Those duplicates are simply ignored. However,
// two different Collectors must not send duplicate descriptors.) This
// method idempotently sends the same descriptors throughout the
// lifetime of the Collector. If a Collector encounters an error while
// executing this method, it must send an invalid descriptor (created
// with NewInvalidDesc) to signal the error to the registry.
//
// Describe implements Describe method of prometheus.Collector interface.
func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	for _, col := range c.collectors() {
		col.Describe(ch)
	}
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent. The
// descriptor of each sent metric is one of those returned by
// Describe. Returned metrics that share the same descriptor must differ
// in their variable label values. This method may be called
// way. Blocking occurs at the expense of total performance of rendering
// concurrently and must therefore be implemented in a concurrency safe
// all registered metrics. Ideally, Collector implementations support
// concurrent readers.
//
// For exp: openebs_volume_uptime{"volName":"", "casType":""} 0
//          openebs_volume_uptime{"volName":"vol1", "casType":"jiva"} 30
//
// In the above case metrics are unique and not duplicated because of
// the labels. In our case we have several collectors such as gauges and
// counters. Each collectors are unique at any instance.
//
// Collect implements Collect method of prometheus.Collector interface.
func (c *collector) Collect(ch chan<- prometheus.Metric) {

	var (
		err         error
		volumeStats v1.VolumeStats
		stats       stats
	)

	c.Lock()
	if c.isRequestInProgress() {
		c.targetRejectRequestCounter.Inc()
		c.targetRejectRequestCounter.Collect(ch)
		c.Unlock()
		return
	}

	c.request = true
	c.Unlock()

	klog.V(2).Info("Get metrics")
	metrics := &c.metrics
	if volumeStats, err = c.get(); err != nil {
		klog.Errorln(err)
		c.setError(err)
	}

	klog.V(2).Info("Parse metrics")
	stats = c.parse(volumeStats, metrics)

	c.set(stats)
	// collect the metrics extracted by collect method
	klog.V(2).Info("Collect metrics")
	for _, col := range c.collectors() {
		col.Collect(ch)
	}
	c.setRequestToFalse()
}

func (c *collector) setError(err error) {
	c.connectionErrorCounter.Inc()
	if _, ok := err.(*colErr); ok {
		c.connectionRetryCounter.Inc()
	}
}

// set is used to set the values gathered from cas to
// prometheus gauges and counters.
func (c *collector) set(volStats stats) {
	//	var replicaAddress, replicaMode strings.Builder

	if !volStats.got {
		klog.Warningf("%s", "setting up empty values")
	}
	c.reads.Set(volStats.reads)
	c.totalReadTime.Set(volStats.totalReadTime)
	c.writes.Set(volStats.writes)
	c.totalWriteTime.Set(volStats.totalWriteTime)
	c.totalReadBytes.Set(volStats.totalReadBytes)
	c.totalWriteBytes.Set(volStats.totalWriteBytes)
	c.totalReadBlockCount.Set(volStats.totalReadBlockCount)
	c.totalWriteBlockCount.Set(volStats.totalWriteBlockCount)
	c.sectorSize.Set(volStats.sectorSize)
	c.logicalSize.Set(volStats.logicalSize)
	c.actualUsed.Set(volStats.actualSize)
	c.sizeOfVolume.Set(volStats.size)

	c.volumeUpTime.WithLabelValues(
		volStats.name,
		volStats.casType,
	).Set(volStats.uptime)

	//	volStats.buildStringof(&replicaAddress, &replicaMode)
	volStats.getReplicaCount()
	c.totalReplicaCounter.Set(volStats.totalReplicaCount)
	c.degradedReplicaCounter.Set(volStats.degradedReplicaCount)
	c.healthyReplicaCounter.Set(volStats.healthyReplicaCount)

	c.volumeStatus.Set(float64(volStats.getVolumeStatus()))
	c.isClientConnected.Set(volStats.isClientConnected)

	return
}
