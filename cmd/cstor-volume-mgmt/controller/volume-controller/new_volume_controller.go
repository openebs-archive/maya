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

package volumecontroller

import (
	"os"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"github.com/openebs/maya/cmd/cstor-volume-mgmt/controller/common"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	openebsScheme "github.com/openebs/maya/pkg/client/generated/clientset/versioned/scheme"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
)

const volumeControllerName = "CStorVolume"

// CStorVolumeController is the controller implementation for CStorVolume resources.
type CStorVolumeController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	// cStorVolumeSynced is used for caches sync to get populated
	cStorVolumeSynced cache.InformerSynced

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

// NewCStorVolumeController returns a new instance of CStorVolume controller
func NewCStorVolumeController(
	kubeclientset kubernetes.Interface,
	clientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	cStorInformerFactory informers.SharedInformerFactory) *CStorVolumeController {

	// obtain references to shared index informers for the cStorVolume resources
	cStorVolumeInformer := cStorInformerFactory.Openebs().V1alpha1().CStorVolumes()

	openebsScheme.AddToScheme(scheme.Scheme)

	// Create event broadcaster to receive events and send them to any EventSink, watcher, or log.
	// Add NewCstorVolumeController types to the default Kubernetes Scheme so Events can be
	// logged for CstorVolume Controller types.
	glog.Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)

	// StartEventWatcher starts sending events received from this EventBroadcaster to the given
	// event handler function. The return value can be ignored or used to stop recording, if
	// desired. Events("") denotes empty namespace
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: os.Getenv("POD_NAME"), Host: os.Getenv("NODE_NAME")})

	controller := &CStorVolumeController{
		kubeclientset:     kubeclientset,
		clientset:         clientset,
		cStorVolumeSynced: cStorVolumeInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "CStorVolume"),
		recorder:          recorder,
	}

	glog.Info("Setting up event handlers")

	// Instantiating QueueLoad before entering workqueue.
	q := common.QueueLoad{}

	// Set up an event handler for when CstorVolume resources change.
	cStorVolumeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if !IsValidCStorVolumeMgmt(obj.(*apis.CStorVolume)) {
				return
			}
			q.Operation = common.QOpAdd
			glog.Infof("Add event received for cstorvolume : %s", obj.(*apis.CStorVolume).Name)
			controller.enqueueCStorVolume(obj.(*apis.CStorVolume), q)
		},
		UpdateFunc: func(old, new interface{}) {
			newCStorVolume := new.(*apis.CStorVolume)
			oldCStorVolume := old.(*apis.CStorVolume)
			if !IsValidCStorVolumeMgmt(newCStorVolume) {
				return
			}
			// Periodic resync will send update events for all known CStorVolume.
			// Two different versions of the same CStorVolume will always have different RVs.
			if newCStorVolume.ResourceVersion == oldCStorVolume.ResourceVersion {
				q.Operation = common.QOpPeriodicSync
				glog.V(4).Infof("cStorVolume periodic sync event : %v %v", newCStorVolume.ObjectMeta.Name, string(newCStorVolume.ObjectMeta.UID))
			} else if IsOnlyStatusChange(oldCStorVolume, newCStorVolume) {
				glog.V(4).Infof("Only cStorVolume status change: %v, %v", newCStorVolume.ObjectMeta.Name, string(newCStorVolume.ObjectMeta.UID))
				return
			} else if IsDestroyEvent(newCStorVolume) {
				q.Operation = common.QOpDestroy
				glog.Infof("cStorVolume Destroy event : %v, %v", newCStorVolume.ObjectMeta.Name, string(newCStorVolume.ObjectMeta.UID))
				controller.recorder.Event(newCStorVolume, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.MessageDestroySynced))
			} else {
				q.Operation = common.QOpModify
				glog.Infof("cStorVolume Modify event : %v, %v", newCStorVolume.ObjectMeta.Name, string(newCStorVolume.ObjectMeta.UID))
				controller.recorder.Event(newCStorVolume, corev1.EventTypeNormal, string(common.SuccessSynced), string(common.MessageModifySynced))
			}
			controller.enqueueCStorVolume(newCStorVolume, q)
		},
		DeleteFunc: func(obj interface{}) {
			glog.Infof("Delete event received for cstorvolume : %s", obj.(*apis.CStorVolume).Name)
		},
	})

	return controller
}
