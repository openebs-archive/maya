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

	clientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"

	"github.com/golang/glog"
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

// Webhook implements a validating webhook.
type Webhook struct {
	//  Server defines parameters for running an golang HTTP server.
	Server *http.Server

	// kubeClient is a standard kubernetes clientset
	KubeClient kubernetes.Interface

	// Clientset is a openebs custom resource package generated for custom API group.
	Clientset clientset.Interface
}

// WParameters are server configures parameters
type WParameters struct {
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

// NewWebhook creates a new instance of a webhook.
func NewWebhook(p WParameters, kubeClient kubernetes.Interface, openebsClient clientset.Interface) (*Webhook, error) {

	pair, err := tls.LoadX509KeyPair(p.CertFile, p.KeyFile)
	if err != nil {
		return nil, err
	}
	wh := &Webhook{
		Server: &http.Server{
			Addr:      fmt.Sprintf(":%v", p.Port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
		KubeClient: kubeClient,
		Clientset:  openebsClient,
	}
	return wh, nil
}

func admissionRequired(ignoredList []string, metadata *metav1.ObjectMeta) bool {
	// skip special kubernetes system namespaces
	for _, namespace := range ignoredList {
		if metadata.Namespace == namespace {
			glog.Infof("Skip validation for %v for it's in special namespace:%v", metadata.Name, metadata.Namespace)
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
func (wh *Webhook) validate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	response := &v1beta1.AdmissionResponse{}

	// validates only if requested operation is DELETE and request kind is PVC
	if ar.Request.Operation != v1beta1.Delete && req.Kind.Kind != "PersistentVolumeClaim" {
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
	pvc, err := wh.KubeClient.CoreV1().PersistentVolumeClaims(req.Namespace).Get(req.Name, metav1.GetOptions{})
	if err != nil {
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("error retrieving PVC: %v", err.Error()),
		}
		return response
	}

	if !validationRequired(ignoredNamespaces, &pvc.ObjectMeta) {
		glog.Infof("Skipping validation for %s/%s due to policy check", pvc.Namespace, pvc.Name)
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	response.Allowed = true
	cStorVolumes, err := wh.Clientset.OpenebsV1alpha1().CStorVolumes("openebs").List(metav1.ListOptions{})
	if err != nil {
		response.Allowed = false
		response.Result = &metav1.Status{
			Message: fmt.Sprintf("error retrieving CstorVolumes: %v", err.Error()),
		}
		return response
	}

	// get the all CStorVolumes resources in openebs namespace(default namespace
	// for CStorvolume resource) to check the source-volume annotation to find
	// out if there is any clone volume exists.
	// if source-volume annotation matches with name of PVC, failed the pvc
	// deletion operation.
	for _, cstorvolume := range cStorVolumes.Items {
		if cstorvolume.Annotations["openebs.io/source-volume"] == pvc.Spec.VolumeName {
			response.Allowed = false
			response.Result = &metav1.Status{
				Status: metav1.StatusFailure, Code: http.StatusForbidden, Reason: "PVC has cloned volume",
				Message: fmt.Sprintf("pvc %q has one or more cloned volume(s)", pvc.Name),
			}
			break
		}
	}
	return response
}

// Serve method for webhook server, handles http requests for webhooks
func (wh *Webhook) Serve(w http.ResponseWriter, r *http.Request) {
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
	glog.Infof("Ready to write reponse ...")
	if _, err := w.Write(resp); err != nil {
		glog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
