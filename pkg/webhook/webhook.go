/*
Copyright 2019 The OpenEBS Authors.

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

package webhook

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	snapclient "github.com/openebs/maya/pkg/client/generated/openebs.io/snapshot/v1alpha1/clientset/internalclientset"
	"k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// (https://github.com/kubernetes/kubernetes/issues/57982)
	defaulter = runtime.ObjectDefaulter(runtimeScheme)
)

// Skip validation in special namespaces, i.e. in kube-system and kube-public
// namespaces the validation will be skipped
var (
	ignoredNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}
	snapshotAnnotation = "snapshot.alpha.kubernetes.io/snapshot"
)

// webhook implements a validating webhook.
type webhook struct {
	//  Server defines parameters for running an golang HTTP server.
	Server *http.Server

	// kubeClient is a standard kubernetes clientset
	kubeClient kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	// snapClientSet is a snaphot custom resource package generated from custom API group.
	snapClientSet snapclient.Interface
}

// Parameters are server configures parameters
type Parameters struct {
	// Port is webhook server port
	Port int
	//CertFile is path to the x509 certificate for https
	CertFile string
	//KeyFile is path to the x509 private key matching `CertFile`
	KeyFile string
}

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
	// defaulting with webhooks:
	// https://github.com/kubernetes/kubernetes/issues/57982
	_ = v1.AddToScheme(runtimeScheme)
}

// New creates a new instance of a webhook.
func New(p Parameters, kubeClient kubernetes.Interface, openebsClient clientset.Interface, snapClient snapclient.Interface) (*webhook, error) {

	pair, err := tls.LoadX509KeyPair(p.CertFile, p.KeyFile)
	if err != nil {
		return nil, err
	}
	wh := &webhook{
		Server: &http.Server{
			Addr:      fmt.Sprintf(":%v", p.Port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
		kubeClient:    kubeClient,
		clientset:     openebsClient,
		snapClientSet: snapClient,
	}
	return wh, nil
}

func admissionRequired(ignoredList []string, metadata *metav1.ObjectMeta) bool {
	// skip special kubernetes system namespaces
	for _, namespace := range ignoredList {
		if metadata.Namespace == namespace {
			glog.V(4).Infof("Skip validation for %v for it's in special namespace:%v", metadata.Name, metadata.Namespace)
			return false
		}
	}
	return true
}

func validationRequired(ignoredList []string, metadata *metav1.ObjectMeta) bool {
	required := admissionRequired(ignoredList, metadata)
	glog.V(4).Infof("Validation policy for %v/%v: required:%v", metadata.Namespace, metadata.Name, required)
	return required
}

// validatePVCDeleteRequest validates the persistentvolumeclaim(PVC) delete request
func (wh *webhook) validatePVCDeleteRequest(req *v1beta1.AdmissionRequest) *v1beta1.AdmissionResponse {
	response := &v1beta1.AdmissionResponse{}
	response.Allowed = true

	// ignore the Delete request of PVC if resource name is empty which
	// can happen as part of cleanup process of namespace
	if req.Name == "" {
		return response
	}

	glog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)

	// TODO* use raw object once validation webhooks support DELETE request
	// object as non nil value https://github.com/kubernetes/kubernetes/issues/66536
	//var pvc corev1.PersistentVolumeClaim
	//err := json.Unmarshal(req.Object.Raw, &pvc)
	//if err != nil {
	//	glog.Errorf("Could not unmarshal raw object: %v, %v", err, req.Object.Raw)
	//	status.Allowed = false
	//	status.Result = &metav1.Status{
	//		Status: metav1.StatusFailure, Code: http.StatusBadRequest, Reason: metav1.StatusReasonBadRequest,
	//		Message: err.Error(),
	//	}
	//	return status
	//}

	// fetch the pvc specifications
	pvc, err := wh.kubeClient.CoreV1().PersistentVolumeClaims(req.Namespace).Get(req.Name, metav1.GetOptions{})
	if err != nil {
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("error retrieving PVC: %v", err.Error()),
		}
		return response
	}

	if !validationRequired(ignoredNamespaces, &pvc.ObjectMeta) {
		glog.V(4).Infof("Skipping validation for %s/%s due to policy check", pvc.Namespace, pvc.Name)
		return response
	}

	// construct source-volume label to list all the matched cstorVolumes
	label := fmt.Sprintf("openebs.io/source-volume=%s", pvc.Spec.VolumeName)
	listOptions := metav1.ListOptions{
		LabelSelector: label,
	}

	// get the all CStorVolumes resources in all namespaces based on the
	// source-volume label to verify if there is any clone volume exists.
	// if source-volume label matches with name of PV, failed the pvc
	// deletion operation.

	cStorVolumes, err := wh.getCstorVolumes(listOptions)
	if err != nil {
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("error retrieving CstorVolumes: %v", err.Error()),
		}
		return response
	}

	if len(cStorVolumes.Items) != 0 {
		response.Allowed = false
		response.Result = &metav1.Status{
			Status: metav1.StatusFailure, Code: http.StatusForbidden, Reason: "PVC with cloned volumes can't be deleted",
			Message: fmt.Sprintf("pvc %q has '%v' cloned volume(s)", pvc.Name, len(cStorVolumes.Items)),
		}
		return response
	}
	return response
}

// validatePVCCreateRequest validates persistentvolumeclaim(PVC) create request
func (wh *webhook) validatePVCCreateRequest(req *v1beta1.AdmissionRequest) *v1beta1.AdmissionResponse {
	response := &v1beta1.AdmissionResponse{}
	response.Allowed = true

	var pvc corev1.PersistentVolumeClaim
	err := json.Unmarshal(req.Object.Raw, &pvc)
	if err != nil {
		glog.Errorf("Could not unmarshal raw object: %v, %v", err, req.Object.Raw)
		response.Allowed = false
		response.Result = &metav1.Status{
			Status:  metav1.StatusFailure,
			Code:    http.StatusBadRequest,
			Reason:  metav1.StatusReasonBadRequest,
			Message: err.Error(),
		}
		return response
	}

	// If snapshot.alpha.kubernetes.io/snapshot annotation represents the clone pvc
	// create request
	snapname := pvc.Annotations[snapshotAnnotation]
	if len(snapname) == 0 {
		return response
	}

	glog.V(4).Infof("AdmissionReview for creating a clone volume Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)
	// get the snapshot object to get snapshotdata object
	// Note: If snapname is empty then below call will retrun error
	snapObj, err := wh.snapClientSet.OpenebsV1alpha1().VolumeSnapshots(pvc.Namespace).Get(snapname, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("failed to get the snapshot object for snapshot name: '%s' namespace: '%s' PVC: '%s'"+
			"error: '%v'", snapname, pvc.Namespace, pvc.Name, err)
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("Failed to get the snapshot object for snapshot name: '%s' namespace: '%s' "+
				"error: '%v'", snapname, pvc.Namespace, err.Error()),
		}
		return response
	}

	snapDataName := snapObj.Spec.SnapshotDataName
	if len(snapDataName) == 0 {
		glog.Errorf("Snapshotdata name is empty for snapshot: '%s' snapshot Namespace: '%s' PVC: '%s'",
			snapname, snapObj.ObjectMeta.Namespace, pvc.Name)
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("Snapshotdata name is empty for snapshot: '%s' snapshot Namespace: '%s'",
				snapname, snapObj.ObjectMeta.Namespace),
		}
		return response
	}
	glog.V(4).Infof("snapshotdata name: '%s'", snapDataName)

	// get the snapDataObj to get the snapshotdataname
	// Note: If snapDataName is empty then below call will return error
	snapDataObj, err := wh.snapClientSet.OpenebsV1alpha1().VolumeSnapshotDatas().Get(snapDataName, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("Failed to get the snapshotdata object for snapshotdata  name: '%s' "+
			"snapName: '%s' namespace: '%s' PVC: '%s' error: '%v'", snapDataName, snapname, snapObj.ObjectMeta.Namespace, pvc.Name, err)
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("Failed to get the snapshotdata object for snapshotdata  name: '%s' "+
				"snapName: '%s' namespace: '%s' error: '%v'", snapDataName, snapname, snapObj.ObjectMeta.Namespace, err.Error()),
		}
		return response
	}

	snapSizeString := snapDataObj.Spec.OpenEBSSnapshot.Capacity
	// If snapshotdata object doesn't consist Capacity field then we will log it and return false.
	if len(snapSizeString) == 0 {
		glog.Infof("snapshot size not found for snapshot name: '%s' snapshot namespace: '%s' snapshotdata name: '%s'",
			snapname, snapObj.ObjectMeta.Namespace, snapDataName)
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("PVC: '%s' creation requires upgrade of volumesnapshotdata name: '%s'", pvc.ObjectMeta.Name, snapDataName),
		}
		return response
	}

	snapCapacity := resource.MustParse(snapSizeString)
	pvcSize := pvc.Spec.Resources.Requests[corev1.ResourceName(corev1.ResourceStorage)]
	if pvcSize.Cmp(snapCapacity) != 0 {
		glog.Errorf("Requested pvc size not matched the snapshot size '%s' belongs to snapshot name: '%s' "+
			"snapshot Namespace: '%s' VolumeSnapshotData '%s'", snapSizeString, snapObj.ObjectMeta.Name, snapObj.ObjectMeta.Namespace, snapDataName)
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("Requested pvc size must be equal to snapshot size '%s' "+
				"which belongs to snapshot name: '%s' snapshot NameSpace: '%s' volumesnapshotdata: '%s'",
				snapSizeString, snapObj.ObjectMeta.Name, snapObj.ObjectMeta.Namespace, snapDataName),
		}
		return response
	}
	return response
}

// validate validates the persistentvolumeclaim(PVC) create, delete request
func (wh *webhook) validate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	response := &v1beta1.AdmissionResponse{}
	response.Allowed = true

	// validates only if request kind is pvc
	if req.Kind.Kind != "PersistentVolumeClaim" {
		return response
	}
	// validates only if requested operation is CREATE or DELETE
	if req.Operation == v1beta1.Create {
		return wh.validatePVCCreateRequest(req)
	} else if req.Operation == v1beta1.Delete {
		return wh.validatePVCDeleteRequest(req)
	}
	return response
}

// getCstorVolumes gets the list of CstorVolumes based in the source-volume labels
func (wh *webhook) getCstorVolumes(listOptions metav1.ListOptions) (*v1alpha1.CStorVolumeList, error) {
	var cStorVolumes *v1alpha1.CStorVolumeList
	var err error

	cStorVolumes, err = wh.clientset.OpenebsV1alpha1().CStorVolumes("").List(listOptions)
	return cStorVolumes, err
}

// Serve method for webhook server, handles http requests for webhooks
func (wh *webhook) Serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		glog.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		glog.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		glog.Errorf("Can't decode body: %v", err)
		admissionResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		if r.URL.Path == "/validate" {
			admissionResponse = wh.validate(&ar)
		}
	}

	admissionReview := v1beta1.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		glog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	glog.V(5).Infof("Ready to write reponse ...")
	if _, err := w.Write(resp); err != nil {
		glog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
