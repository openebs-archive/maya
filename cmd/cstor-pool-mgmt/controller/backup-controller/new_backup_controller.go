/*
Copyright 2018 The OpenEBS Authors.

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

package backupcontroller

import (
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"

	//clientset "github.com/openebs/maya/pkg/client/clientset/versioned"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	//openebsScheme "github.com/openebs/maya/pkg/client/clientset/versioned/scheme"
	openebsScheme "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset/scheme"
	//informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	informers "github.com/openebs/maya/pkg/client/generated/informer/externalversions"
)

const backupControllerName = "CStorBackup"

// CStorBackupController is the controller implementation for CStorBackup resources.
type CStorBackupController struct {
	// kubeclientset is a standard kubernetes clientset.
	kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	// CStorBackupSynced is used for caches sync to get populated
	CStorBackupSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewCStorBackupController returns a new cStor Replica controller instance
func NewCStorBackupController(
	kubeclientset kubernetes.Interface,
	clientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	cStorInformerFactory informers.SharedInformerFactory) *CStorBackupController {

	// obtain references to shared index informers for the CStorBackup resources.
	CStorBackupInformer := cStorInformerFactory.Openebs().V1alpha1().CStorBackups()

	openebsScheme.AddToScheme(scheme.Scheme)

	// Create event broadcaster
	// Add cStor-Replica-controller types to the default Kubernetes Scheme so Events can be
	// logged for cStor-Replica-controller types.
	glog.V(4).Info("Creating backup event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)

	// StartEventWatcher starts sending events received from this EventBroadcaster to the given
	// event handler function. The return value can be ignored or used to stop recording, if
	// desired. Events("") denotes empty namespace
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: backupControllerName})

	controller := &CStorBackupController{
		kubeclientset:     kubeclientset,
		clientset:         clientset,
		CStorBackupSynced: CStorBackupInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "CStorBackup"),
		recorder:          recorder,
	}

	glog.Info("Setting up event handlers for backup")

	// Instantiating QueueLoad before entering workqueue.
	q := common.QueueLoad{}

	// Set up an event handler for when cStorReplica resources change.
	CStorBackupInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// ToDo : Enqueue object for processing in case of added event
			// Note: AddFunc is called when a new object comes into etcd
			// Note : In case controller reboots and existing object in etcd can cause delivery of
			// add event when the controller comes again. Be careful in this part and handle accordingly.
			csb := obj.(*apis.CStorBackup)

			if !IsRightCStorPoolMgmt(csb) {
				return
			}

			if IsDeletionFailedBefore(csb) || IsErrorDuplicate(csb) {
				return
			}
			q.Operation = common.QOpAdd
			glog.Infof("cStorBackup Added event : %v, %v", csb.ObjectMeta.Name, string(csb.ObjectMeta.UID))
			controller.recorder.Event(csb, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.MessageCreateSynced))
			csb.Status.Phase = apis.CSBStatusPending
			csb, _ = controller.clientset.OpenebsV1alpha1().CStorBackups(csb.Namespace).Update(csb)
			csbData := create_csb_data(csb)

			_, err := controller.clientset.OpenebsV1alpha1().CStorBackupDatas(csb.Namespace).Create(csbData)
			if err != nil {
				glog.Errorf("Failed to create backupdata: error '%s'", err.Error())
				return
			}
			controller.enqueueCStorBackup(csb, q)
		},
		UpdateFunc: func(old, new interface{}) {
			newCSB := new.(*apis.CStorBackup)
			oldCSB := old.(*apis.CStorBackup)
			glog.Infof("CStorBackup sync done %s %s", newCSB.Spec.Name, newCSB.Spec.IncrementalBackupName)
			if !IsRightCStorPoolMgmt(newCSB) {
				return
			}
			if IsOnlyStatusChange(oldCSB, newCSB) {
				glog.Infof("Only cSB status change: %v, %v", newCSB.ObjectMeta.Name, string(newCSB.ObjectMeta.UID))
				return
			}
			if IsDeletionFailedBefore(newCSB) || IsErrorDuplicate(newCSB) {
				return
			}
			if newCSB.ResourceVersion == oldCSB.ResourceVersion {
				q.Operation = common.QOpSync
				glog.Infof("CstorBackup status sync event for %s", newCSB.ObjectMeta.Name)
				glog.Infof("CStorBackup sync done %s", newCSB.Spec.IncrementalBackupName)
				controller.recorder.Event(newCSB, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.StatusSynced))
			} else if IsDestroyEvent(newCSB) {
				q.Operation = common.QOpDestroy
				glog.Infof("cStorBAckup Destroy event : %v, %v", newCSB.ObjectMeta.Name, string(newCSB.ObjectMeta.UID))
				controller.recorder.Event(newCSB, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.MessageDestroySynced))
			} else {
				q.Operation = common.QOpModify
				listOptions := v1.ListOptions{}
				var backupData *apis.CStorBackupData
				backupDataList, err := controller.clientset.OpenebsV1alpha1().CStorBackupDatas(newCSB.Namespace).List(listOptions)
				for _, bkpData := range backupDataList.Items {
					if bkpData.Spec.Name == newCSB.Spec.Name {
						backupData = &bkpData
					}
				}
				if backupData == nil {
					glog.Infof("Failed to find backupData for %s", newCSB.Spec.Name)
					csbData := create_csb_data(newCSB)
					_, err := controller.clientset.OpenebsV1alpha1().CStorBackupDatas(newCSB.Namespace).Create(csbData)
					if err != nil {
						glog.Errorf("Failed to create backupdata: error '%s'", err.Error())
						return
					}
					glog.Infof("Successfully Created backupData for %s", newCSB.Spec.Name, csbData.Spec.IncrementalBackupName)
				} else {
					update_csb_data(newCSB, backupData)
					backupData, err = controller.clientset.OpenebsV1alpha1().CStorBackupDatas(newCSB.Namespace).Update(backupData)
					if err != nil {
						glog.Errorf("Failed to update backupdata: error '%s'", err.Error())

					}
					glog.Infof("Successfully Updated backupData for %s %s", newCSB.Spec.IncrementalBackupName, backupData.Spec.IncrementalBackupName)
				}
				glog.Infof("cStorBackup Modify event : %v, %v", newCSB.ObjectMeta.Name, string(newCSB.ObjectMeta.UID))
				controller.recorder.Event(newCSB, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.MessageModifySynced))
			}
			//controller.enqueueCStorBackup(newCSB, q)
			// ToDo : Enqueue object for processing in case of update event
			// Note : UpdateFunc is called in following three cases:
			// 1. When object is updated/patched i.e. Resource version of object changes.
			// 2. When object is deleted i.e. the deletion timestap of object is set.
			// 3. After every resync interval.
		},
		DeleteFunc: func(obj interface{}) {
			csb := obj.(*apis.CStorBackup)
			if !IsRightCStorPoolMgmt(csb) {
				return
			}
			glog.Infof("csb Resource deleted event: %v, %v", csb.ObjectMeta.Name, string(csb.ObjectMeta.UID))

		},
	})
	return controller
}

// enqueueCStorBackup takes a CStorBackup resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than CStorBackup.
func (c *CStorBackupController) enqueueCStorBackup(obj *apis.CStorBackup, q common.QueueLoad) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	q.Key = key
	c.workqueue.AddRateLimited(q)
}

func create_csb_data(csb *apis.CStorBackup) *apis.CStorBackupData {
	backupData := &apis.CStorBackupData{}
	backupData.Name = csb.Name
	backupData.Spec.Name = csb.Spec.Name
	backupData.Spec.VolumeName = csb.Spec.VolumeName
	backupData.Spec.LastSnapshotName = csb.Spec.LastSnapshotName
	backupData.Spec.IncrementalBackupName = csb.Spec.IncrementalBackupName
	return backupData
}

func update_csb_data(csb *apis.CStorBackup, backupData *apis.CStorBackupData) {
	backupData.Spec.LastSnapshotName = csb.Spec.LastSnapshotName
	backupData.Spec.IncrementalBackupName = csb.Spec.IncrementalBackupName
}
