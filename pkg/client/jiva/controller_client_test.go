package client

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/openebs/maya/types/v1"

	"github.com/openebs/maya/pkg/util"

	utiltesting "k8s.io/client-go/util/testing"
)

var (
	replicaResponse    = `{"actions":{},"id":"1","links":{"self":"http://10.44.0.2:9502/v1/replicas/1"},"replicacounter":2,"revisioncounter":"0","type":"replica"}`
	controllerResponse = `{"Name":"vol1","ReadIOPS":"0","ReplicaCounter":0,"RevisionCounter":0,"SCSIIOCount":{},"SectorSize":"4096","Size":"1073741824","TotalReadBlockCount":"0","TotalReadTime":"0","TotalWriteTime":"0","TotalWriteBlockCount":"0","UpTime":158.667823193,"UsedBlocks":"5","UsedLogicalBlocks":"0","WriteIOPS":"0","actions":{},"links":{"self":"http://10.42.0.1:9501/v1/stats"},"type":"stats"}`
	v1ReplicasResponse = `{"createTypes":{"replica":"http://10.1.2.17:9501/v1/replicas"},"data":[{"actions":{"preparerebuild":"http://10.1.2.17:9501/v1/replicas/dGNwOi8vMTAuMS4xLjk6OTUwMg==?action=preparerebuild","verifyrebuild":"http://10.1.2.17:9501/v1/replicas/dGNwOi8vMTAuMS4xLjk6OTUwMg==?action=verifyrebuild"},"address":"tcp://10.1.1.9:9502","id":"dGNwOi8vMTAuMS4xLjk6OTUwMg==","links":{"self":"http://10.1.2.17:9501/v1/replicas/dGNwOi8vMTAuMS4xLjk6OTUwMg=="},"mode":"RW","type":"replica"},{"actions":{"preparerebuild":"http://10.1.2.17:9501/v1/replicas/dGNwOi8vMTAuMS4yLjE4Ojk1MDI=?action=preparerebuild","verifyrebuild":"http://10.1.2.17:9501/v1/replicas/dGNwOi8vMTAuMS4yLjE4Ojk1MDI=?action=verifyrebuild"},"address":"tcp://10.1.2.18:9502","id":"dGNwOi8vMTAuMS4yLjE4Ojk1MDI=","links":{"self":"http://10.1.2.17:9501/v1/replicas/dGNwOi8vMTAuMS4yLjE4Ojk1MDI="},"mode":"RW","type":"replica"}],"links":{"self":"http://10.1.2.17:9501/v1/replicas"},"resourceType":"replica","type":"collection"}`
)

func TestGetVolumeStats(t *testing.T) {
	var (
		replicaClient    *ReplicaClient
		controllerClient *ControllerClient
		replicaStatus    v1.VolStatus
		controllerStatus v1.VolumeMetrics
	)

	tests := map[string]struct {
		fakeHandler utiltesting.FakeHandler
		err         error
	}{
		"200OK": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(replicaResponse),
				T:            t,
			},
			err: nil,
		},
		"500InternalServerError": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   500,
				ResponseBody: string(replicaResponse),
				T:            t,
			},
			err: util.ErrInternalServerError,
		},
		"503ServerUnavailable": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   503,
				ResponseBody: string(replicaResponse),
				T:            t,
			},
			err: util.ErrServerUnavailable,
		},
		"BadRequest": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode: 400,
				T:          t,
			},
			err: io.EOF,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			defer server.Close()
			_, err := replicaClient.GetVolumeStats(server.URL, &replicaStatus)
			if err != tt.err {
				t.Errorf("GetVolumeStats(%v) => got %v, want %v", server.URL, err, tt.err)
			}
		})
	}
	tests = map[string]struct {
		fakeHandler utiltesting.FakeHandler
		err         error
	}{
		"200_OK": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(controllerResponse),
				T:            t,
			},
			err: nil,
		},
		"500_InternalServerError": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   500,
				ResponseBody: string(controllerResponse),
				T:            t,
			},
			err: util.ErrInternalServerError,
		},
		"503_ServerUnavailable": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   503,
				ResponseBody: string(controllerResponse),
				T:            t,
			},
			err: util.ErrServerUnavailable,
		},
		"Bad_Request": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode: 400,
				T:          t,
			},
			err: io.EOF,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			defer server.Close()
			_, err := controllerClient.GetVolumeStats(server.URL, "/stats", &controllerStatus)
			if err != tt.err {
				t.Errorf("GetVolumeStats(%v) => got %v, want %v", server.URL, err, tt.err)
			}
		})
	}

}

func TestGetVolumeAccessMode(t *testing.T) {
	var (
		status           = ReplicaCollection{}
		controllerClient *ControllerClient
	)
	tests := map[string]struct {
		fakeHandler utiltesting.FakeHandler
		err         error
	}{
		"200OK": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   200,
				ResponseBody: string(v1ReplicasResponse),
				T:            t,
			},
			err: nil,
		},
		"500InternalServerError": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   500,
				ResponseBody: string(v1ReplicasResponse),
				T:            t,
			},
			err: util.ErrInternalServerError,
		},
		"503ServerUnavailable": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode:   503,
				ResponseBody: string(v1ReplicasResponse),
				T:            t,
			},
			err: util.ErrServerUnavailable,
		},
		"BadRequest": {
			fakeHandler: utiltesting.FakeHandler{
				StatusCode: 400,
				T:          t,
			},
			err: io.EOF,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(&tt.fakeHandler)
			defer server.Close()
			_, err := controllerClient.GetVolumeStats(server.URL, "/replicas", &status)
			if err != tt.err {
				t.Errorf("GetVolumeAccessMode(%v, %v) => got %v, want %v", server.URL, status, err, tt.err)
			}
		})
	}
}
