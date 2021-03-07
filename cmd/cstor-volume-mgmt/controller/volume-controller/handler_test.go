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

package volumecontroller

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/openebs/maya/cmd/cstor-volume-mgmt/controller/common"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
	"github.com/openebs/maya/pkg/client/k8s"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func fakeStrToQuantity(capacity string) resource.Quantity {
	qntCapacity, _ := resource.ParseQuantity(capacity)
	return qntCapacity
}

// TestGetVolumeResource checks if volume resource created is successfully got.
func TestGetVolumeResource(t *testing.T) {
	fakeKubeClient := fake.NewSimpleClientset()
	fakeOpenebsClient := openebsFakeClientset.NewSimpleClientset()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(fakeKubeClient, time.Second*30)
	openebsInformerFactory := informers.NewSharedInformerFactory(fakeOpenebsClient, time.Second*30)

	// Instantiate the cStor Volume controllers.
	volumeController := NewCStorVolumeController(fakeKubeClient, fakeOpenebsClient, kubeInformerFactory,
		openebsInformerFactory)

	testVolumeResource := map[string]struct {
		expectedVolumeName string
		test               *apis.CStorVolume
	}{
		"img1VolumeResource": {
			expectedVolumeName: "abc",
			test: &apis.CStorVolume{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "volume1",
					UID:       types.UID("abc"),
					Namespace: string(common.DefaultNameSpace),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP: "0.0.0.0",
					Capacity: fakeStrToQuantity("5G"),
					Status:   "init",
				},
				Status: apis.CStorVolumeStatus{},
			},
		},
		"img2VolumeResource": {
			expectedVolumeName: "abcd",
			test: &apis.CStorVolume{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "volume2",
					UID:       types.UID("abcd"),
					Namespace: string(common.DefaultNameSpace),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP: "0.0.0.0",
					Capacity: fakeStrToQuantity("15G"),
					Status:   "init",
				},
				Status: apis.CStorVolumeStatus{},
			},
		},
	}
	for desc, ut := range testVolumeResource {
		// Create Volume resource
		_, err := volumeController.clientset.OpenebsV1alpha1().CStorVolumes(string(common.DefaultNameSpace)).
			Create(context.TODO(), ut.test, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Desc:%v, Unable to create resource : %v", desc, ut.test.ObjectMeta.Name)
		}
		// Get the created volume resource using name
		cStorVolumeObtained, err := volumeController.getVolumeResource(ut.test.ObjectMeta.Name)
		if string(cStorVolumeObtained.ObjectMeta.UID) != ut.expectedVolumeName {
			t.Fatalf("Desc:%v, VolumeName mismatch, Expected:%v, Got:%v", desc, ut.expectedVolumeName,
				string(cStorVolumeObtained.ObjectMeta.UID))
		}
	}
}

// TestIsValidCStorVolumeMgmt is to check if right sidecar does operation with env match.
func TestIsValidCStorVolumeMgmt(t *testing.T) {
	testVolumeResource := map[string]struct {
		expectedOutput bool
		test           *apis.CStorVolume
	}{
		"img2VolumeResource": {
			expectedOutput: true,
			test: &apis.CStorVolume{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:       "volume2",
					UID:        types.UID("abcd"),
					Finalizers: []string{"cstorvolume.openebs.io/finalizer"},
					Namespace:  string(common.DefaultNameSpace),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP: "0.0.0.0",
					Capacity: fakeStrToQuantity("15G"),
					Status:   "init",
				},
				Status: apis.CStorVolumeStatus{},
			},
		},
	}
	for desc, ut := range testVolumeResource {
		os.Setenv("OPENEBS_IO_CSTOR_VOLUME_ID", string(ut.test.UID))
		obtainedOutput := IsValidCStorVolumeMgmt(ut.test)
		if obtainedOutput != ut.expectedOutput {
			t.Fatalf("Desc:%v, Expected:%v, Got:%v", desc, ut.expectedOutput,
				obtainedOutput)
		}
		os.Setenv("OPENEBS_IO_CSTOR_VOLUME_ID", "")
	}
}

// TestIsValidCStorVolumeMgmtNegative is to check if right sidecar does operation with env match.
func TestIsValidCStorVolumeMgmtNegative(t *testing.T) {
	testVolumeResource := map[string]struct {
		expectedOutput bool
		test           *apis.CStorVolume
	}{
		"img2VolumeResource": {
			expectedOutput: false,
			test: &apis.CStorVolume{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:       "volume2",
					UID:        types.UID("abcd"),
					Finalizers: []string{"cstorvolume.openebs.io/finalizer"},
					Namespace:  string(common.DefaultNameSpace),
				},
				Spec: apis.CStorVolumeSpec{
					TargetIP: "0.0.0.0",
					Capacity: fakeStrToQuantity("15G"),
					Status:   "init",
				},
				Status: apis.CStorVolumeStatus{},
			},
		},
	}
	for desc, ut := range testVolumeResource {
		os.Setenv("OPENEBS_IO_CSTOR_VOLUME_ID", string("awer"))
		obtainedOutput := IsValidCStorVolumeMgmt(ut.test)
		if obtainedOutput != ut.expectedOutput {
			t.Fatalf("Desc:%v, Expected:%v, Got:%v", desc, ut.expectedOutput,
				obtainedOutput)
		}
		os.Setenv("OPENEBS_IO_CSTOR_VOLUME_ID", "")
	}
}

func TestCreateEventObj(t *testing.T) {
	tests := map[string]struct {
		cstorVolume       *apis.CStorVolume
		event             *v1.Event
		podName, nodeName string
		clientset         kubernetes.Interface
	}{
		"event": {
			cstorVolume: &apis.CStorVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "csv-1",
					UID:             types.UID("abcd"),
					Namespace:       string(common.DefaultNameSpace),
					ResourceVersion: "1111",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "CstorVolume",
					APIVersion: "v1alpha1",
				},
				Status: apis.CStorVolumeStatus{
					Phase: apis.CStorVolumePhase(common.CVStatusHealthy),
				},
			},
			event: &v1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "csv-1.Healthy",
					Namespace: string(common.DefaultNameSpace),
				},
				InvolvedObject: v1.ObjectReference{
					Kind:            string(k8s.CStorVolumeCRKK),
					APIVersion:      string(k8s.OEV1alpha1KA),
					Name:            "csv-1",
					Namespace:       string(common.DefaultNameSpace),
					UID:             types.UID("abcd"),
					ResourceVersion: "1111",
				},
				FirstTimestamp: metav1.Time{Time: time.Now()},
				LastTimestamp:  metav1.Time{Time: time.Now()},
				Count:          1,
				Message:        fmt.Sprintf(common.EventMsgFormatter, "Healthy"),
				Reason:         "Healthy",
				Type:           getEventType(common.CStorVolumeStatus("Healthy")),
				Source: v1.EventSource{
					Component: "mypod",
					Host:      "mynode",
				},
			},
			podName:  "mypod",
			nodeName: "mynode",

			clientset: fake.NewSimpleClientset(),
		},
	}

	for desc, ut := range tests {
		t.Run(desc, func(t *testing.T) {
			os.Setenv("POD_NAME", ut.podName)
			os.Setenv("NODE_NAME", ut.nodeName)
			cvController := CStorVolumeController{kubeclientset: ut.clientset}
			event := cvController.createEventObj(ut.cstorVolume)
			if !reflect.DeepEqual(event.ObjectMeta, ut.event.ObjectMeta) {
				t.Errorf("Failed to create event, invalid ObjectMeta: want=%v got=%v", ut.event.ObjectMeta, event.ObjectMeta)
			}
			if !reflect.DeepEqual(event.InvolvedObject, ut.event.InvolvedObject) {
				t.Errorf("Failed to create event, invalid InvolvedObject: want=%v got=%v", ut.event.InvolvedObject, event.InvolvedObject)
			}
			if !reflect.DeepEqual(event.Count, ut.event.Count) {
				t.Errorf("Failed to create event, invalid Count: want=%v got=%v", ut.event.Count, event.Count)
			}
			if !reflect.DeepEqual(event.Message, ut.event.Message) {
				t.Errorf("Failed to create event, invalid Message: want=%v got=%v", ut.event.Message, event.Message)
			}
			if !reflect.DeepEqual(event.Reason, ut.event.Reason) {
				t.Errorf("Failed to create event, invalid Reason: want=%v got=%v", ut.event.Reason, event.Reason)
			}
			if !reflect.DeepEqual(event.Type, ut.event.Type) {
				t.Errorf("Failed to create event, invalid Type: want=%v got=%v", ut.event.Type, event.Type)
			}
			if !reflect.DeepEqual(event.Source, ut.event.Source) {
				t.Errorf("Failed to create event, invalid Source: want=%v got=%v", ut.event.Source, event.Source)
			}
			// = %v, want %v", "got, ut.event
			os.Unsetenv("POD_NAME")
			os.Unsetenv("NODE_NAME")
		})
	}
}

func TestGetEventType(t *testing.T) {
	tests := map[string]struct {
		phase     common.CStorVolumeStatus
		eventType string
	}{
		"Normal event Init":     {phase: common.CVStatusInit, eventType: v1.EventTypeNormal},
		"Normal event Healthy":  {phase: common.CVStatusHealthy, eventType: v1.EventTypeNormal},
		"Normal event Degraded": {phase: common.CVStatusDegraded, eventType: v1.EventTypeNormal},
		"Warning event Error":   {phase: common.CVStatusError, eventType: v1.EventTypeWarning},
		"Warning event Invalid": {phase: common.CVStatusInvalid, eventType: v1.EventTypeWarning},
		"Warning event Offline": {phase: common.CVStatusOffline, eventType: v1.EventTypeWarning},
	}
	for desc, ut := range tests {
		t.Run(desc, func(t *testing.T) {
			if got := getEventType(ut.phase); got != ut.eventType {
				t.Errorf("Incorrect event type= %v, want %v", got, ut.eventType)
			}
		})
	}
}
