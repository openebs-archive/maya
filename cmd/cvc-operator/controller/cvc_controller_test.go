/*
Copyright 2019 The OpenEBS Authors

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

package cstorvolumeclaim

import (
	"reflect"
	"testing"
	"time"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type conditionMergeTestCase struct {
	description    string
	cvc            *apis.CStorVolumeClaim
	newConditions  []apis.CStorVolumeClaimCondition
	finalCondtions []apis.CStorVolumeClaimCondition
}

func TestMergeResizeCondition(t *testing.T) {
	currentTime := metav1.Now()

	cvc := getCVC([]apis.CStorVolumeClaimCondition{
		{
			Type:               apis.CStorVolumeClaimResizing,
			LastTransitionTime: currentTime,
		},
	})

	noConditionCVC := getCVC([]apis.CStorVolumeClaimCondition{})

	conditionFalseTime := metav1.Now()
	newTime := metav1.NewTime(time.Now().Add(1 * time.Hour))

	testCases := []conditionMergeTestCase{
		{
			description:    "when removing all conditions",
			cvc:            cvc.DeepCopy(),
			newConditions:  []apis.CStorVolumeClaimCondition{},
			finalCondtions: []apis.CStorVolumeClaimCondition{},
		},
		{
			description: "adding new condition",
			cvc:         cvc.DeepCopy(),
			newConditions: []apis.CStorVolumeClaimCondition{
				{
					Type: apis.CStorVolumeClaimResizePending,
				},
			},
			finalCondtions: []apis.CStorVolumeClaimCondition{
				{
					Type: apis.CStorVolumeClaimResizePending,
				},
			},
		},
		{
			description: "adding same condition with new timestamp",
			cvc:         cvc.DeepCopy(),
			newConditions: []apis.CStorVolumeClaimCondition{
				{
					Type:               apis.CStorVolumeClaimResizing,
					LastTransitionTime: newTime,
				},
			},
			finalCondtions: []apis.CStorVolumeClaimCondition{
				{
					Type:               apis.CStorVolumeClaimResizing,
					LastTransitionTime: newTime,
				},
			},
		},
		{
			description: "adding same condition but with different status",
			cvc:         cvc.DeepCopy(),
			newConditions: []apis.CStorVolumeClaimCondition{
				{
					Type:               apis.CStorVolumeClaimResizing,
					LastTransitionTime: conditionFalseTime,
				},
			},
			finalCondtions: []apis.CStorVolumeClaimCondition{
				{
					Type:               apis.CStorVolumeClaimResizing,
					LastTransitionTime: conditionFalseTime,
				},
			},
		},
		{
			description: "when no condition exists on pvc",
			cvc:         noConditionCVC.DeepCopy(),
			newConditions: []apis.CStorVolumeClaimCondition{
				{
					Type:               apis.CStorVolumeClaimResizing,
					LastTransitionTime: currentTime,
				},
			},
			finalCondtions: []apis.CStorVolumeClaimCondition{
				{
					Type:               apis.CStorVolumeClaimResizing,
					LastTransitionTime: currentTime,
				},
			},
		},
	}

	for _, testcase := range testCases {
		updateConditions := MergeResizeConditionsOfCVC(testcase.cvc.Status.Conditions, testcase.newConditions)

		if !reflect.DeepEqual(updateConditions, testcase.finalCondtions) {
			t.Errorf("Expected updated conditions for test %s to be %v but got %v",
				testcase.description,
				testcase.finalCondtions, updateConditions)
		}
	}

}

func getCVC(conditions []apis.CStorVolumeClaimCondition) *apis.CStorVolumeClaim {
	cvc := &apis.CStorVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "openebs"},
		Spec: apis.CStorVolumeClaimSpec{
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): resource.MustParse("2Gi"),
			},
		},
		Status: apis.CStorVolumeClaimStatus{
			Phase:      apis.CStorVolumeClaimPhaseBound,
			Conditions: conditions,
			Capacity: v1.ResourceList{
				v1.ResourceStorage: resource.MustParse("2Gi"),
			},
		},
	}
	return cvc
}
