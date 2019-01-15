package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type restoreAPIOps struct {
	req  *http.Request
	resp http.ResponseWriter
}

// restoreV1alpha1SpecificRequest deals with restore API requests
func (s *HTTPServer) restoreV1alpha1SpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	restoreOp := &restoreAPIOps{
		req:  req,
		resp: resp,
	}

	switch req.Method {
	case "POST":
		return restoreOp.create()
	}
	return nil, CodedError(405, ErrInvalidMethod)
}

// Create is http handler which handles restore-create request
func (rOps *restoreAPIOps) create() (interface{}, error) {

	restore := &v1alpha1.CStorRestore{}

	err := decodeBody(rOps.req, restore)
	if err != nil {
		return nil, err
	}

	// namespace is expected
	if len(strings.TrimSpace(restore.Namespace)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create restore '%v': missing namespace", restore.Name))
	}

	// restore name is expected
	if len(strings.TrimSpace(restore.Spec.Name)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create restore: missing restore name "))
	}

	// volume name is expected
	if len(strings.TrimSpace(restore.Spec.VolumeName)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create restore '%v': missing volume name", restore.Name))
	}

	// restoreIP is expected
	if len(strings.TrimSpace(restore.Spec.RestoreSrc)) == 0 {
		return nil, CodedError(400, fmt.Sprintf("failed to create restore '%v': missing restoreSrc", restore.Name))
	}

	openebsClient, _ := loadClientFromServiceAccount()
	/*
		splitName := strings.Split(restore.Spec.Name, "-")
		if len(splitName) >= 2 {
			restore.Name = strings.Join(splitName[0:len(splitName)-1], "-")
		} else {
			restore.Name = restore.Spec.Name
		}
	*/
	//Check if this schedule is already present
	/*
		rstList, err := openebsClient.OpenebsV1alpha1().CStorRestores(restore.Namespace).List(listOptions)
		for _, rst := range rstList.Items {
			if rst.Name == restore.Name {
				rst.Spec.RestoreSrc = restore.Spec.RestoreSrc
				openebsClient.OpenebsV1alpha1().CStorRestores(rst.Namespace).Update(&rst)
				glog.Infof("Creating incremental restore %s volume %s poolUUID:%v",
					restore.Spec.Name,
					restore.Spec.VolumeName, rst.ObjectMeta.Labels["cstorpool.openebs.io/uid"])
				return "", nil
			}
		}
	*/

	//Get List of cvr's related to this pvc
	listOptions := v1.ListOptions{}
	listOptions = v1.ListOptions{
		LabelSelector: "openebs.io/persistent-volume=" + restore.Spec.VolumeName,
	}
	cvrList, err := openebsClient.OpenebsV1alpha1().CStorVolumeReplicas("").List(listOptions)

	//Select a healthy csr for restore
	for _, cvr := range cvrList.Items {
		restore.Name = restore.Spec.Name + cvr.ObjectMeta.Labels["cstorpool.openebs.io/uid"]
		restore.ObjectMeta.Labels = map[string]string{
			"cstorpool.openebs.io/uid": cvr.ObjectMeta.Labels["cstorpool.openebs.io/uid"],
		}
		glog.Infof("Creating restore %s for volume %q poolUUID:%v", restore.Name,
			restore.Spec.VolumeName,
			restore.ObjectMeta.Labels["cstorpool.openebs.io/uid"])
		_, err = openebsClient.OpenebsV1alpha1().CStorRestores(restore.Namespace).Create(restore)
		if err != nil {
			glog.Errorf("Failed to create restore: error '%s'", err.Error())
			return nil, CodedError(500, err.Error())
		}
		glog.Infof("Restore CR created successfully: name '%s'", restore.Name)
	}

	return len(cvrList.Items), nil
}
