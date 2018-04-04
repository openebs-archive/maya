package controller

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/cstorops/pool"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/cstorops/uzfs"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/cstorops/volumereplica"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var poolCrd = `cat <<EOF | sudo kubectl create -f -
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cstorpools.openebs.io
spec:
  group: openebs.io
  names:
    kind: CStorPool
    listKind: CStorPoolList
    plural: cstorpools
    shortNames:
    - cstorpool
  scope: Cluster
  version: v1alpha1`

var clearPoolCrd = `cat <<EOF | sudo kubectl delete -f -
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cstorpools.openebs.io
spec:
  group: openebs.io
  names:
    kind: CStorPool
    listKind: CStorPoolList
    plural: cstorpools
    shortNames:
    - cstorpool
  scope: Cluster
  version: v1alpha1`

var volumeReplicaCrd = `cat <<EOF | sudo kubectl create -f -
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cstorvolumereplicas.openebs.io
spec:
  group: openebs.io
  names:
    kind: CStorVolumeReplica
    listKind: CStorVolumeReplicaList
    plural: cstorvolumereplicas
    shortNames:
    - cvr
  scope: Cluster
  version: v1alpha1
  `
var clearVolumeReplicaCrd = `cat <<EOF | sudo kubectl delete -f -
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cstorvolumereplicas.openebs.io
spec:
  group: openebs.io
  names:
    kind: CStorVolumeReplica
    listKind: CStorVolumeReplicaList
    plural: cstorvolumereplicas
    shortNames:
    - cvr
  scope: Cluster
  version: v1alpha1
  `
var img1CStorPoolResource = `cat <<EOF | sudo kubectl create -f -
apiVersion: openebs.io/v1alpha1
kind: CStorPool
metadata:
  name: pool1
  node: node-host-label
spec:
  disks:
   diskList: ["/tmp/img1.img"]
  poolSpec:
   poolName: pool1
   cacheFile: /tmp/pool1.cache
   poolType: mirror
`

// TestCheckForCStorPoolCRD tests CStorPool crd availability with timeout.
func TestCheckForCStorPoolCRD(t *testing.T) {
	kubeconfig := os.Getenv("HOME") + "/.kube/config"
	cfg, err := getClusterConfig(kubeconfig)
	if err != nil {
		t.Fatalf(err.Error())
	}
	openebsClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("Error building clientset: %s", err.Error())
	}

	done := make(chan bool)

	go func(done chan bool) {
		checkForCStorPoolCRD(openebsClient)
		done <- true
	}(done)

	execShResult(poolCrd)

	select {
	case <-time.After(10 * time.Second):
		t.Fatalf("timeout")
	case <-done:
		execShResult(clearPoolCrd)
	}
}

// TestCheckForCStorVolumeReplicaCRD tests CStorVolumeReplica crd availability with timeout.
func TestCheckForCStorVolumeReplicaCRD(t *testing.T) {
	kubeconfig := os.Getenv("HOME") + "/.kube/config"
	cfg, err := getClusterConfig(kubeconfig)
	if err != nil {
		t.Fatalf(err.Error())
	}
	openebsClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("Error building clientset: %s", err.Error())
	}

	done := make(chan bool)

	go func(done chan bool) {
		checkForCStorVolumeReplicaCRD(openebsClient)
		done <- true
	}(done)

	execShResult(volumeReplicaCrd)

	select {
	case <-time.After(10 * time.Second):
		t.Fatalf("timeout")
	case <-done:
		execShResult(clearVolumeReplicaCrd)
	}
}

// TestApplyCStorPool tests cstor pool resource creation
func TestApplyCStorPool(t *testing.T) {
	execShResult(clearPoolCrd)
	execShResult(clearVolumeReplicaCrd)
	execShResult(poolCrd)
	execShResult(volumeReplicaCrd)
	execShResult(img1CStorPoolResource)

	done := make(chan bool)
	go func() {
		uzfs.CheckForZrepl()
		done <- true
	}()
	select {
	case <-time.After(20 * time.Second):
		t.Fatalf("Timeout error")
	case <-done:
	}

	actualPoolName, err := pool.GetPoolName()
	if err == nil {
		pool.DeletePool(actualPoolName)
	}

	kubeconfig := os.Getenv("HOME") + "/.kube/config"

	cfg, err := getClusterConfig(kubeconfig)
	if err != nil {
		t.Fatalf(err.Error())
	}
	openebsClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("Error building clientset: %s", err.Error())
	}

	var wg sync.WaitGroup
	wg.Add(1)

	time.Sleep(5 * time.Second)
	go func() {
		StartControllers(kubeconfig)
	}()
	time.Sleep(5 * time.Second)

	var poolObtained apis.CStorPool

	go func() {
		poolList, err := openebsClient.OpenebsV1alpha1().CStorPools().List(metav1.ListOptions{})
		if err != nil {
			t.Errorf("Unable to List pool: %v", err)
		}
		for _, poolObtained = range poolList.Items {
			time.Sleep(10 * time.Second)
			createdPool, err := pool.GetPoolName()
			if err != nil {
				t.Errorf("Error : %v", err.Error())
			}
			if createdPool != poolObtained.Spec.PoolSpec.PoolName {
				t.Errorf("Fail : %v is not equal to %v", createdPool, poolObtained.Spec.PoolSpec.PoolName)
			}
		}
		wg.Done()
	}()
	wg.Wait()
	pool.DeletePool(poolObtained.Spec.PoolSpec.PoolName)
	pool.DeletePool(poolObtained.Spec.PoolSpec.PoolName)
}

// TestApplyCStorVolumeReplica tests volume replica creation.
func TestApplyCStorVolumeReplica(t *testing.T) {
	done := make(chan bool)
	go func() {
		uzfs.CheckForZrepl()
		done <- true
	}()
	select {
	case <-time.After(20 * time.Second):
		t.Fatalf("Timeout error")
	case <-done:

	}
	execShResult(clearPoolCrd)
	execShResult(clearVolumeReplicaCrd)
	execShResult(poolCrd)
	execShResult(volumeReplicaCrd)
	execShResult(img1CStorPoolResource)

	actualPoolName, err := pool.GetPoolName()
	if err == nil {
		pool.DeletePool(actualPoolName)
	}

	kubeconfig := os.Getenv("HOME") + "/.kube/config"

	go func() {
		StartControllers(kubeconfig)
	}()
	time.Sleep(5 * time.Second)

	poolName, err := pool.GetPoolName()
	if err != nil {
		t.Fatalf(err.Error())
	}

	tests := map[string]struct {
		expectedVolName string
		actualVolNames  []string
		resourceYaml    string
	}{
		"cStorVolumeReplicaResource_1": {
			expectedVolName: poolName + "/" + "vol1",
			resourceYaml: `cat <<EOF | sudo kubectl create -f -
apiVersion: openebs.io/v1alpha1
kind: CStorVolumeReplica
metadata:
  name: pvc-ee171da3-07d5-11e8-a5be-42010a8001be-cstor-rep-9440ab
  annotations:
  openebs.io/cstor-pool-guid: 7b99e406-1260-11e8-aa43-00505684eb2e

spec:
  cStorControllerIP: 10.210.102.206
  volName: vol1
  capacity: 100MB`,
		},
		"cStorVolumeReplicaResource_2": {
			expectedVolName: poolName + "/" + "vol2",
			resourceYaml: `cat <<EOF | sudo kubectl create -f -
apiVersion: openebs.io/v1alpha1
kind: CStorVolumeReplica
metadata:
  name: pvc-ee171da3-07d5-11e8-a5be-42010a8001be-cstor-rep-9440ac
  annotations:
  openebs.io/cstor-pool-guid: 7b99e406-1260-11e8-aa43-00505684eb2e

spec:
  cStorControllerIP: 10.210.102.206
  volName: vol2
  capacity: 100MB`,
		},
	}

	for desc, ut := range tests {
		_, err := execShResult(ut.resourceYaml)
		if err != nil {
			t.Errorf("Unable to apply cvr resource, %v ", err.Error())
		}
		time.Sleep(5 * time.Second)
		ut.actualVolNames = volumereplica.GetVolumes()
		if err != nil {
			t.Errorf("desc: %v, Error : %v", desc, err.Error())
		}

		var availableFlag = false
		for _, actualVolName := range ut.actualVolNames {
			if actualVolName == ut.expectedVolName {
				availableFlag = true
				break
			}
		}
		if !availableFlag {
			t.Errorf("desc: %v, Fail : %v is not available", desc, ut.expectedVolName)
		}
	}
	execShResult(clearPoolCrd)
	execShResult(clearVolumeReplicaCrd)
	pool.DeletePool(poolName)
	pool.DeletePool(poolName)
}
