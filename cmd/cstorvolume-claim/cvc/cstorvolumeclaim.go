/*
Copyright 2017 The OpenEBS Authors

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

package cvc

import (
	"strconv"

	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	csp "github.com/openebs/maya/pkg/cstor/pool/v1alpha3"
	cv "github.com/openebs/maya/pkg/cstor/volume/v1alpha1"
	cvr "github.com/openebs/maya/pkg/cstor/volumereplica/v1alpha1"
	errors "github.com/openebs/maya/pkg/errors/v1alpha1"
	svc "github.com/openebs/maya/pkg/kubernetes/service/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	//kubeConfigPath = "/home/user/.kube/config"
	ports = []corev1.ServicePort{
		corev1.ServicePort{
			Name:     "cstor-iscsi",
			Port:     3260,
			Protocol: "TCP",
			TargetPort: intstr.IntOrString{
				IntVal: 3260,
			},
		},
		corev1.ServicePort{
			Name:     "cstor-grpc",
			Port:     7777,
			Protocol: "TCP",
			TargetPort: intstr.IntOrString{
				IntVal: 7777,
			},
		},
		corev1.ServicePort{
			Name:     "mgmt",
			Port:     6060,
			Protocol: "TCP",
			TargetPort: intstr.IntOrString{
				IntVal: 6060,
			},
		},
		corev1.ServicePort{
			Name:     "exporter",
			Port:     9500,
			Protocol: "TCP",
			TargetPort: intstr.IntOrString{
				IntVal: 9500,
			},
		},
	}
)

func getStorageClass(storageClassName string,
) (*storagev1.StorageClass, error) {
	if storageClassName == "" {
		return nil, errors.New("failed to get storageclass: storageclass missing")
	}
	scObj, err := sc.NewKubeClient().Get(storageClassName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to get storageclass {%s}",
			storageClassName,
		)
	}
	return scObj, nil
}

func getReplicationFactor(scName string,
) (int, int, error) {

	scObj, err := getStorageClass(scName)
	if err != nil {
		return 0, 0, errors.Wrapf(
			err,
			"failed to get storageclass obj {%s}",
			scName,
		)
	}
	count := scObj.Parameters["replicaCount"]

	rfactor, err := strconv.Atoi(count)
	if err != nil {
		return 0, 0, errors.Wrapf(
			err,
			"failed to convert to int {%s}",
			count,
		)
	}
	glog.Infof("rfactor and confactor %d , %d", rfactor, rfactor/2+1)
	return rfactor, (rfactor/2 + 1), nil

}

func getFromParameters(scName string,
) (int, string, error) {

	scObj, err := getStorageClass(scName)
	if err != nil {
		return 0, "", errors.Wrapf(
			err,
			"failed to get storageclass obj {%s}",
			scName,
		)
	}
	count := scObj.Parameters["replicaCount"]
	rCount, err := strconv.Atoi(count)
	if err != nil {
		return 0, "", errors.Wrapf(
			err,
			"failed to convert to int {%s}",
			count,
		)
	}

	spcName := scObj.Parameters["storagePoolClaim"]
	return rCount, spcName, nil
}

func listCStorPools(
	spcName string,
	replicaCount int,
) (*apis.CStorPoolList, error) {

	if spcName == "" {
		return nil, errors.New("failed to list cstorpool: spc missing")
	}

	labelSelector := "openebs.io/storage-pool-claim=" + spcName

	cstorPoolList, err := csp.KubeClient().List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to list cstorpool for spc {%s}",
			spcName,
		)
	}
	if len(cstorPoolList.Items) < replicaCount {
		return nil, errors.New("not enough pools available to create replicas")
	}
	return cstorPoolList, nil
}

func createTargetService(storageClassName, pvName string,
) (*corev1.Service, error) {

	labels := map[string]string{
		"openebs.io/target-service":      "cstor-target-svc",
		"openebs.io/storage-engine-type": "cstor",
		"openebs.io/persistent-volume":   pvName,
	}
	selectors := map[string]string{
		"app":                          "cstor-volume-manager",
		"openebs.io/target":            "cstor-target",
		"openebs.io/persistent-volume": pvName,
	}
	annotations := map[string]string{
		"openebs.io/storage-class-ref": "name: " + storageClassName,
	}

	svcObj, err := svc.NewKubeClient(svc.WithNamespace("openebs")).Get(pvName, metav1.GetOptions{})
	if k8serror.IsNotFound(err) {
		svcObj, err = svc.NewBuilder().
			WithName(pvName).
			WithLabelsNew(labels).
			WithAnnotations(annotations).
			WithSelectorsNew(selectors).
			WithPorts(ports).
			Build()

		svcObj, err = svc.NewKubeClient(svc.WithNamespace("openebs")).Create(svcObj)
	}
	return svcObj, err
}

func createCStorVolumecr(service *corev1.Service,
	claim *apis.CStorVolumeClaim,
	scName string,
) (*apis.CStorVolume, error) {

	labels := map[string]string{
		"openebs.io/target-service":    "cstor-target-svc",
		"openebs.io/persistent-volume": claim.Name,
	}
	annotations := map[string]string{
		//"openebs.io/storage-class-ref": "name: " + storageClassName,
		"openebs.io/fs-type": "ext4",
		"openebs.io/lun":     "0",
	}

	qCap := claim.Spec.Capacity[corev1.ResourceStorage]
	rfactor, cfactor, err := getReplicationFactor(scName)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to get replica-count and factor {%s}",
			scName,
		)
	}

	cvObj, err := cv.NewKubeclient(cv.WithNamespace("openebs")).Get(claim.Name, metav1.GetOptions{})
	if k8serror.IsNotFound(err) {
		cvObj, err := cv.NewBuilder().
			WithName(claim.Name).
			WithLabelsNew(labels).
			WithAnnotationsNew(annotations).
			WithTargetIP(service.Spec.ClusterIP).
			WithCapacity(qCap.String()).
			WithIQN(claim.Name).
			WithTargetPortal(service.Spec.ClusterIP + ":" + "3260").
			WithTargetPort("3260").
			WithReplicationFactor(rfactor).
			WithConsistencyFactor(cfactor).
			Build()
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"failed to get cstorvolume {%v}",
				cvObj,
			)
		}

		return cv.NewKubeclient(cv.WithNamespace("openebs")).Create(cvObj)
	}
	return cvObj, err
}

func createCStorVolumeReplica(
	service *corev1.Service,
	volume *apis.CStorVolume,
	scName string,
) (*apis.CStorVolumeReplica, error) {

	replicaCount, spcName, err := getFromParameters(scName)
	if err != nil {
		return nil, err
	}

	poolList, err := listCStorPools(spcName, replicaCount)
	if err != nil {
		return nil, err
	}

	if len(poolList.Items) < replicaCount {
		return nil, errors.Wrapf(
			err,
			"not enough pools to provision expected count {%d} actual count {%d}",
			replicaCount,
			len(poolList.Items),
		)
	}

	for i, pool := range poolList.Items {
		if i < replicaCount {
			err := creatCVR(service, volume, &pool)
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, err
}

func creatCVR(
	service *corev1.Service,
	volume *apis.CStorVolume,
	pool *apis.CStorPool,
) error {

	labels := map[string]string{
		"cstorpool.openebs.io/name":    pool.Name,
		"cstorpool.openebs.io/uid":     string(pool.UID),
		"cstorvolume.openebs.io/name":  volume.Name,
		"openebs.io/persistent-volume": volume.Name,
	}
	annotations := map[string]string{
		//cstorpool.openebs.io/hostname: {{ pluck .ListItems.currentRepeatResource .ListItems.cvolPoolNodeList.pools | first }}
		"cstorpool.openebs.io/hostname": pool.Labels["kubernetes.io/hostname"],
	}
	finalizer := []string{
		"cstorvolumereplica.openebs.io/finalizer",
	}

	cvrObj, err := cvr.NewBuilder().
		WithName(volume.Name + "-" + pool.Name).
		WithLabelsNew(labels).
		WithAnnotationsNew(annotations).
		WithFinalizers(finalizer).
		WithTargetIP(service.Spec.ClusterIP).
		WithCapacity(volume.Spec.Capacity).
		Build()
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to build cstorvolumereplica {%v}",
			cvrObj,
		)
	}
	_, err = cvr.NewKubeclient(cvr.WithNamespace("openebs")).Create(cvrObj)
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to create cstorvolumereplica {%v}",
			cvrObj,
		)
	}
	return nil
}
