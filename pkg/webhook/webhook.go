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
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	"k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

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
)

// webhook implements a validating webhook.
type webhook struct {
	//  Server defines parameters for running an golang HTTP server.
	Server *http.Server

	// kubeClient is a standard kubernetes clientset
	kubeClient kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface
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
func New(p Parameters, kubeClient kubernetes.Interface, openebsClient clientset.Interface) (*webhook, error) {

	pair, err := tls.LoadX509KeyPair(p.CertFile, p.KeyFile)
	if err != nil {
		return nil, err
	}
	wh := &webhook{
		Server: &http.Server{
			Addr:      fmt.Sprintf(":%v", p.Port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
		kubeClient: kubeClient,
		clientset:  openebsClient,
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

// validate validates the persistentvolumeclaim(PVC) delete request
func (wh *webhook) validate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	response := &v1beta1.AdmissionResponse{}

	// validates only if requested operation is DELETE and request kind is PVC
	if ar.Request.Operation != v1beta1.Delete || req.Kind.Kind != "PersistentVolumeClaim" {
		response.Allowed = true
		return response
	}

	// ignore the Delete request of PVC if resource name is empty which
	// can happen as part of cleanup process of namespace
	if ar.Request.Name == "" {
		response.Allowed = true
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
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
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

	response.Allowed = true
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
