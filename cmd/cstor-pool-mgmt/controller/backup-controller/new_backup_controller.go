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
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	"github.com/openebs/maya/cmd/cstor-pool-mgmt/controller/common"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"

	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	openebsScheme "github.com/openebs/maya/pkg/client/generated/clientset/versioned/scheme"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
)

const backupControllerName = "CStorBackup"

// BackupController is the controller implementation for CStorBackup resources.
type BackupController struct {
	// kubeclientset is a standard kubernetes clientset.
	kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	// BackupSynced is used for caches sync to get populated
	BackupSynced cache.InformerSynced

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

// NewCStorBackupController returns a new cStor Backup controller instance
func NewCStorBackupController(
	kubeclientset kubernetes.Interface,
	clientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	cStorInformerFactory informers.SharedInformerFactory) *BackupController {

	// obtain references to shared index informers for the CStorBackup resources.
	BackupInformer := cStorInformerFactory.Openebs().V1alpha1().CStorBackups()

	err := openebsScheme.AddToScheme(scheme.Scheme)
	if err != nil {
		klog.Fatalf("Error adding scheme to openebs scheme: %s", err.Error())
		return nil
	}

	// Create event broadcaster
	// Add backup-cstor-controller types to the default Kubernetes Scheme so Events can be
	// logged for backup-cstor-controller types.
	klog.V(4).Info("Creating backup event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)

	// StartEventWatcher starts sending events received from this EventBroadcaster to the given
	// event handler function. The return value can be ignored or used to stop recording, if
	// desired. Events("") denotes empty namespace
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: backupControllerName})

	controller := &BackupController{
		kubeclientset: kubeclientset,
		clientset:     clientset,
		BackupSynced:  BackupInformer.Informer().HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "CStorBackup"),
		recorder:      recorder,
	}

	klog.Info("Setting up event handlers for backup")

	// Clean any pending backup for this cstor pool
	controller.cleanupOldBackup(clientset)

	// Instantiating QueueLoad before entering workqueue.
	q := common.QueueLoad{}

	// Set up an event handler for when CStorBackup resources change.
	BackupInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// Note: AddFunc is called when a new object comes into etcd
			// Note : In case controller reboots and existing object in etcd can cause delivery of
			// add event when the controller comes again. Be careful in this part and handle accordingly.
			bkp := obj.(*apis.CStorBackup)

			if !IsRightCStorPoolMgmt(bkp) {
				return
			}
			controller.handleBKPAddEvent(bkp, &q)
		},
		UpdateFunc: func(oldVar, newVar interface{}) {
			// Note : UpdateFunc is called in following three cases:
			// 1. When object is updated/patched i.e. Resource version of object changes.
			// 2. When object is deleted i.e. the deletion timestamp of object is set.
			// 3. After every resync interval.
			newbkp := newVar.(*apis.CStorBackup)
			oldbkp := oldVar.(*apis.CStorBackup)

			if !IsRightCStorPoolMgmt(newbkp) {
				return
			}

			controller.handleBKPUpdateEvent(oldbkp, newbkp, &q)
		},
		DeleteFunc: func(obj interface{}) {
			bkp := obj.(*apis.CStorBackup)
			if !IsRightCStorPoolMgmt(bkp) {
				return
			}
			klog.Infof("CStorBackup Resource delete event: %v, %v", bkp.ObjectMeta.Name, string(bkp.ObjectMeta.UID))
		},
	})
	return controller
}

// enqueueCStorBackup takes a CStorBackup resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than CStorBackup.
func (c *BackupController) enqueueCStorBackup(obj *apis.CStorBackup, q common.QueueLoad) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	q.Key = key
	c.workqueue.AddRateLimited(q)
}

// cleanupOldBackup set fail status to old pending backup
func (c *BackupController) cleanupOldBackup(clientset clientset.Interface) {
	bkplabel := "cstorpool.openebs.io/uid=" + os.Getenv(string(common.OpenEBSIOCStorID))
	bkplistop := metav1.ListOptions{
		LabelSelector: bkplabel,
	}
	bkplist, err := clientset.OpenebsV1alpha1().CStorBackups(metav1.NamespaceAll).List(bkplistop)
	if err != nil {
		return
	}

	for _, bkp := range bkplist.Items {
		switch bkp.Status {
		case apis.BKPCStorStatusInProgress:
			//Backup was in in-progress state
			laststat := findLastBackupStat(clientset, bkp)
			updateBackupStatus(clientset, bkp, laststat)
		case apis.BKPCStorStatusDone:
			continue
		default:
			//Set backup status as failed
			updateBackupStatus(clientset, bkp, apis.BKPCStorStatusFailed)
		}
	}
}

// updateBackupStatus will update the backup status to given status
func updateBackupStatus(clientset clientset.Interface, bkp apis.CStorBackup, status apis.CStorBackupStatus) {
	bkp.Status = status

	_, err := clientset.OpenebsV1alpha1().CStorBackups(bkp.Namespace).Update(&bkp)
	if err != nil {
		klog.Errorf("Failed to update backup(%s) status(%s)", status, bkp.Name)
		return
	}
}

// findLastBackupStat will find the status of backup from last completed-backup
func findLastBackupStat(clientset clientset.Interface, bkp apis.CStorBackup) apis.CStorBackupStatus {
	lastbkpname := bkp.Spec.BackupName + "-" + bkp.Spec.VolumeName
	lastbkp, err := clientset.OpenebsV1alpha1().CStorCompletedBackups(bkp.Namespace).Get(lastbkpname, v1.GetOptions{})
	if err != nil {
		// Unable to fetch the last backup, so we will return fail state
		klog.Errorf("Failed to fetch last completed backup:%s error:%s", lastbkpname, err.Error())
		return apis.BKPCStorStatusFailed
	}

	// let's check if snapname matches with current snapshot name
	if bkp.Spec.SnapName == lastbkp.Spec.SnapName || bkp.Spec.SnapName == lastbkp.Spec.PrevSnapName {
		return apis.BKPCStorStatusDone
	}

	// lastbackup snap/prevsnap doesn't match with bkp snapname
	return apis.BKPCStorStatusFailed
}

// handleBKPAddEvent is to handle add operation of backup controller
func (c *BackupController) handleBKPAddEvent(bkp *apis.CStorBackup, q *common.QueueLoad) {
	q.Operation = common.QOpAdd
	klog.Infof("CStorBackup event added: %v, %v", bkp.ObjectMeta.Name, string(bkp.ObjectMeta.UID))
	c.recorder.Event(bkp, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.MessageCreateSynced))
	c.enqueueCStorBackup(bkp, *q)
}

// handleBKPUpdateEvent is to handle add operation of backup controller
func (c *BackupController) handleBKPUpdateEvent(oldbkp, newbkp *apis.CStorBackup, q *common.QueueLoad) {
	klog.V(2).Infof("Received Update for backup:%s", oldbkp.ObjectMeta.Name)

	if newbkp.ResourceVersion == oldbkp.ResourceVersion {
		return
	}

	if IsDestroyEvent(newbkp) {
		q.Operation = common.QOpDestroy
		klog.Infof("CStorBackup Destroy event : %v, %v", newbkp.ObjectMeta.Name, string(newbkp.ObjectMeta.UID))
		c.recorder.Event(newbkp, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.MessageDestroySynced))
	} else {
		klog.Infof("CStorBackup Modify event : %v, %v", newbkp.ObjectMeta.Name, string(newbkp.ObjectMeta.UID))
		q.Operation = common.QOpSync
		c.recorder.Event(newbkp, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.MessageModifySynced))
	}
	c.enqueueCStorBackup(newbkp, *q)
}
