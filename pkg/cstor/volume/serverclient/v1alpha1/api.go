// Copyright © 2018-2019 The OpenEBS Authors
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

package v1alpha1

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	"github.com/openebs/maya/pkg/client/generated/cstor-volume-mgmt/v1alpha1"
	"github.com/openebs/maya/pkg/util"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
)

// sudo $ISTGTCONTROL snapdestroy vol1 snapname1 0
// sudo $ISTGTCONTROL snapcreate vol1 snapname1

// constants
const (
	VolumeGrpcListenPort = 7777
	CmdSnapCreate        = "SNAPCREATE"
	CmdSnapDestroy       = "SNAPDESTROY"
	CmdSnapList          = "SNAPLIST"
	//IoWaitTime is the time interval for which the IO has to be stopped before doing snapshot operation
	IoWaitTime = 10
	//TotalWaitTime is the max time duration to wait for doing snapshot operation on all the replicas
	TotalWaitTime   = 60
	ProtocolVersion = 1
)

//CommandStatus is the response from istgt for control commands
type CommandStatus struct {
	Response string `json:"response"`
}

//APIUnixSockVar is unix socker variable
var APIUnixSockVar util.UnixSock

// Server represents the gRPC server
type Server struct {
	Client clientset.Interface
}

type snapshotInfo struct {
	Name  string `json:"name"`
	Props struct {
		Logicalreferenced int64  `json:"logicalreferenced,string"`
		Compressratio     string `json:"compressratio"`
		Used              int64  `json:"used,string"`
		Referenced        int64  `json:"referenced,string"`
		Written           int64  `json:"written,string"`
	} `json:"properties"`
}

type snaplist struct {
	Snapshots []struct {
		ReplicaID string         `json:"replica_id"`
		Snapshot  []snapshotInfo `json:"snapshot"`
	} `json:"snapshots"`
}

func init() {
	APIUnixSockVar = util.RealUnixSock{}
}

// RunVolumeSnapCreateCommand performs snapshot create operation and sends back the response
func (s *Server) RunVolumeSnapCreateCommand(ctx context.Context, in *v1alpha1.VolumeSnapCreateRequest) (*v1alpha1.VolumeSnapCreateResponse, error) {
	klog.Infof("Received snapshot create request. volname = %s, snapname = %s, version = %d", in.Volume, in.Snapname, in.Version)
	volcmd, err := CreateSnapshot(ctx, in)
	if err == nil {
		err = s.addSnapInfoToCVR(in.Volume, in.Snapname)
		if err != nil {
			klog.Errorf("failed to update CVR's snapshot err=%s", err)
			// destroy the created snapshot
			DeleteSnapshot(ctx, &v1alpha1.VolumeSnapDeleteRequest{Volume: in.Volume, Snapname: in.Snapname})
		}
	}
	return volcmd, err

}

// RunVolumeSnapDeleteCommand performs snapshot create operation and sends back the response
func (s *Server) RunVolumeSnapDeleteCommand(ctx context.Context, in *v1alpha1.VolumeSnapDeleteRequest) (*v1alpha1.VolumeSnapDeleteResponse, error) {
	klog.Infof("Received snapshot delete request. volname = %s, snapname = %s, version = %d", in.Volume, in.Snapname, in.Version)
	volcmd, err := DeleteSnapshot(ctx, in)
	if err == nil {
		err = s.removeSnapInfoFromCVR(in.Volume, in.Snapname)
		if err != nil {
			klog.Errorf("failed to clean CVR's snapshot err=%s", err)
		}
	}

	return volcmd, err
}

// CreateSnapshot sends snapcreate command to istgt and returns the response
func CreateSnapshot(ctx context.Context, in *v1alpha1.VolumeSnapCreateRequest) (*v1alpha1.VolumeSnapCreateResponse, error) {
	sockresp, err := APIUnixSockVar.SendCommand(fmt.Sprintf("%s %s %s %v %v",
		CmdSnapCreate, in.Volume, in.Snapname, IoWaitTime, TotalWaitTime))
	respstr := "ERR"
	if nil != sockresp && len(sockresp) > 1 {
		respstr = sockresp[1]
	}
	status := CommandStatus{
		Response: respstr,
	}
	jsonresp, _ := json.Marshal(status)
	resp := &v1alpha1.VolumeSnapCreateResponse{
		Status: jsonresp,
	}
	return resp, err
}

// DeleteSnapshot sends snapdelete command to istgt and returns the response
func DeleteSnapshot(ctx context.Context, in *v1alpha1.VolumeSnapDeleteRequest) (*v1alpha1.VolumeSnapDeleteResponse, error) {
	sockresp, err := APIUnixSockVar.SendCommand(fmt.Sprintf("%s %s %s %v %v",
		CmdSnapDestroy, in.Volume, in.Snapname, IoWaitTime, TotalWaitTime))
	respstr := "ERR"
	if nil != sockresp && len(sockresp) > 1 {
		respstr = sockresp[1]
	}
	status := CommandStatus{
		Response: respstr,
	}
	jsonresp, _ := json.Marshal(status)
	resp := &v1alpha1.VolumeSnapDeleteResponse{
		Status: jsonresp,
	}
	return resp, err
}

// addSnapInfoToCVR updates the relevant CVR with snapshot details
func (s *Server) addSnapInfoToCVR(volume, snapshot string) error {
	cvrlist, err := getCVRList(s.Client, volume)
	if err != nil {
		return errors.Errorf("failed to fetch CVR list err=%s", err)
	}
	snaps, err := fetchSnapshotProperties(volume, snapshot)
	if err != nil {
		return err
	}

	for _, snap := range snaps.Snapshots {
		if len(snap.Snapshot) != 1 {
			klog.Warningf("invalid count=%d of snapshot for replica=%s",
				len(snap.Snapshot), snap.ReplicaID)
			continue
		}
		rsnap := parseSnapshot(snap.Snapshot)
		cvr, err := getCVRFromReplicaID(cvrlist, snap.ReplicaID)
		if err != nil {
			return err
		}

		err = updateCVRWithSnapshot(s.Client, &cvr, rsnap, true)
		if err != nil {
			return err
		}
	}

	return nil
}

// removeSnapInfoFromCVR removes the snapshot details from relevant CVR
func (s *Server) removeSnapInfoFromCVR(volume, snapshot string) error {
	cvrlist, err := getCVRList(s.Client, volume)
	if err != nil {
		return errors.Errorf("failed to fetch CVR list")
	}

	snapinfo := map[string]apis.CStorSnapshotInfo{
		snapshot: apis.CStorSnapshotInfo{},
	}
	for _, cvr := range cvrlist.Items {
		err = updateCVRWithSnapshot(s.Client, &cvr, snapinfo, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateCVRWithSnapshot(client clientset.Interface, cvr *apis.CStorVolumeReplica, snap map[string]apis.CStorSnapshotInfo, isAdd bool) error {
	retryCount := 0
retry:
	if retryCount == 5 {
		return errors.Errorf("update attempts exhausted for cvr=%s", cvr.Name)
	}
	isDirty := false
	for snapname, snapinfo := range snap {
		if _, ok := cvr.Status.Snapshots[snapname]; !ok && isAdd {
			cvr.Status.Snapshots[snapname] = snapinfo
			isDirty = true
		} else if !isAdd {
			delete(cvr.Status.Snapshots, snapname)
			delete(cvr.Status.PendingSnapshots, snapname)
			isDirty = true
		}
	}

	if !isDirty {
		return nil
	}

	_, err := client.OpenebsV1alpha1().CStorVolumeReplicas(cvr.Namespace).Update(cvr)
	if err != nil {
		if k8serror.IsConflict(err) {
			var gerr error
			retryCount = retryCount + 1
			cvr, gerr = client.OpenebsV1alpha1().CStorVolumeReplicas(cvr.Namespace).Get(cvr.Name, v1.GetOptions{})
			if err != nil {
				return errors.Errorf("failed to get updated cvr=%s err=%s", cvr.Name, gerr)
			}
			klog.Warningf("failed to update cvr=%s err=%s retryCount=%v", cvr.Name, err, retryCount)
			goto retry
		}
		klog.Errorf("failed to update cvr=%s err=%s", cvr.Name, err)
		return err
	}
	return nil
}

func getCVRFromReplicaID(cvrlist *apis.CStorVolumeReplicaList, replicaid string) (apis.CStorVolumeReplica, error) {
	for _, cvr := range cvrlist.Items {
		if cvr.Spec.ReplicaID == replicaid {
			return cvr, nil
		}
	}
	return apis.CStorVolumeReplica{}, errors.Errorf("no CVR found for replica=%v", replicaid)
}

func parseSnapshot(snaplist []snapshotInfo) map[string]apis.CStorSnapshotInfo {
	smap := map[string]apis.CStorSnapshotInfo{}

	for _, v := range snaplist {
		cratio, err := strconv.ParseInt(v.Props.Compressratio, 10, 64)
		if err != nil {
			klog.Warningf("failed to convert compressratio=%s for snap=%s. using default=0 err=%s", v.Props.Compressratio, v.Name, err)
			cratio = 0
		}
		cratioString := fmt.Sprintf("%.2fx", float64(cratio)/100)
		rsnap := apis.CStorSnapshotInfo{
			LogicalReferenced: v.Props.Logicalreferenced,
			Written:           v.Props.Written,
			Compression:       cratioString,
			Referenced:        v.Props.Referenced,
			Used:              v.Props.Used,
		}

		smap[v.Name] = rsnap
	}
	return smap
}

func getCVRList(client clientset.Interface, volume string) (*apis.CStorVolumeReplicaList, error) {
	listOptions := v1.ListOptions{
		LabelSelector: "openebs.io/persistent-volume=" + volume,
	}

	return client.OpenebsV1alpha1().CStorVolumeReplicas("").List(listOptions)
}

func fetchSnapshotProperties(volume, snapshot string) (snaplist, error) {
	snaps := snaplist{}
	sockresp, err := APIUnixSockVar.SendCommand(fmt.Sprintf("%s %s@%s %v %v",
		CmdSnapList, volume, snapshot, IoWaitTime, TotalWaitTime))
	if err != nil {
		return snaps, errors.Errorf("failed to get snapshot list.. err=%s", err)
	}
	if sockresp == nil {
		return snaps, errors.New("no response recieved for snaplist")
	}

	for _, v := range sockresp {
		if strings.Contains(v, "ERR") || strings.Contains(v, "ERROR") {
			return snaps, fmt.Errorf("snaplist returns error %v", sockresp)
		}
	}

	for _, v := range sockresp {
		if strings.Contains(v, "SNAPLIST") {
			s := strings.TrimPrefix(v, "SNAPLIST ")
			err := json.Unmarshal([]byte(s), &snaps)
			if err != nil {
				return snaps, errors.Errorf("failed to parse snapshotlist json err=%s data=%s", err, s)
			}
			break
		}
	}

	if (len(snaps.Snapshots)) == 0 {
		return snaps, errors.New("failed to fetch snapshot list")
	}

	return snaps, nil
}
