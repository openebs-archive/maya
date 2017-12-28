package mcrd

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/openebs/maya/cmd/maya-nodebot/storage/block"
	"github.com/openebs/maya/cmd/maya-nodebot/types/v1"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/clientset/versioned"
	crdscheme "github.com/openebs/maya/pkg/client/clientset/versioned/scheme"
	informers "github.com/openebs/maya/pkg/client/informers/externalversions"
	listers "github.com/openebs/maya/pkg/client/listers/openebs/v1alpha1"
	//samplev1alpha1 "github.com/testsamplecontroller/sample-controller/temp"
)

const controllerAgentName = "maya-nodebot"

const (
	// SuccessSynced is used as part of the Event 'reason' when a spc is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a spc fails
	// to sync due to a spc of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a spc already existing
	MessageResourceExists = "Resource %q already exists and is not managed by spc"
	// MessageResourceSynced is the message used for an Event fired when a spc
	// is synced successfully
	MessageResourceSynced = "SPC synced successfully"
)

// Controller is the controller implementation for spc resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	clientset clientset.Interface

	spcLister listers.StoragePoolClaimLister

	spcSynced cache.InformerSynced

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

type QueueLoad struct {
	key       string
	operation string
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	clientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	spcInformerFactory informers.SharedInformerFactory) *Controller {

	// obtain references to shared index informers for the sp and spc
	// types.
	spcInformer := spcInformerFactory.Openebs().V1alpha1().StoragePoolClaims()

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	crdscheme.AddToScheme(scheme.Scheme)

	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset: kubeclientset,
		clientset:     clientset,
		spcLister:     spcInformer.Lister(),
		spcSynced:     spcInformer.Informer().HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "spcs"),
		recorder:      recorder,
	}

	glog.Info("Setting up event handlers")
	// Set up an event handler for when spc resources change
	q := QueueLoad{}
	spcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			q.operation = "add"
			controller.enqueueSpc(obj, q)
		},
		UpdateFunc: func(old, new interface{}) {
			newspc := new.(*apis.StoragePoolClaim)
			oldspc := old.(*apis.StoragePoolClaim)
			if newspc.ResourceVersion == oldspc.ResourceVersion {
				// Periodic resync will send update events for all known spc.
				// Two different versions of the same spc will always have different RVs.
				return
			}
			q.operation = "update"
			controller.enqueueSpc(new, q)
		},
		DeleteFunc: func(obj interface{}) {
			q.operation = "delete"
			controller.enqueueSpc(obj, q)
		},
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting spc controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.spcSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch two workers to process spc resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var q QueueLoad
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if q, ok = obj.(QueueLoad); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// spc resource to be synced.
		if err := c.syncHandler(q.key, q.operation); err != nil {
			return fmt.Errorf("error syncing '%s': %s", q.key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", q.key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the spcUpdated resource
// with the current status of the resource.
func (c *Controller) syncHandler(key, operation string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the spcUpdated resource with this namespace/name
	if namespace == "" {
		namespace = "default"
	}

	//spcUpdated, err := c.spcLister.StoragePoolClaims(namespace).Get(name)
	spcUpdated, err := c.clientset.OpenebsV1alpha1().StoragePoolClaims(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		// The spc resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("spcUpdated '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	var sp apis.StoragePool
	flag := IsDiskAvailable(spcUpdated.Spec.Name)
	if flag == false {
		runtime.HandleError(fmt.Errorf("Disk not available: %s", spcUpdated.Spec.Name))
		return nil
	}

	switch operation {
	case "add":
		sp = DiskOperations(spcUpdated, namespace)
		Createsp(sp, c, namespace)
		break
	case "update":
		//IsDiskBeingUsed needs to be implemeneted
		//IsDiskUnusedMounted needs to be implemeneted
		sp = DiskOperations(spcUpdated, namespace)
		Updatesp(sp, c, namespace)
		break
	case "delete":
		//IsDiskBeingUsed needs to be implemeneted
		//IsDiskUnusedMounted needs to be implemeneted
		err := block.UnMount(spcUpdated.Spec.Name)
		if err != nil {
			runtime.HandleError(fmt.Errorf("Unable to unmount ", err))
			break
		}
		Deletesp(spcUpdated.Spec.Name, c, namespace)
		break
	}

	return nil
}

// enqueueSpc takes a spc resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than spc.
func (c *Controller) enqueueSpc(obj interface{}, q QueueLoad) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	q.key = key
	c.workqueue.AddRateLimited(q)
}

func GetNodeName() (string, error) {
	var Nodename string
	NodenameByte, err := exec.Command("uname", "-n").Output()
	Nodename = string(NodenameByte)
	return Nodename, err
}

func DiskOperations(spc *apis.StoragePoolClaim, namespace string) apis.StoragePool {
	Message := ""

	Nodename, err := GetNodeName()
	if err != nil {
		Message = Message + "unable to get nodename " + err.Error()
		//util.CheckErr(err, util.Fatal)
	}

	res, err := block.Format(spc.Spec.Name, spc.Spec.Format)
	if err != nil {
		Message = Message + " unable to format" + err.Error()
		//util.CheckErr(err, util.Fatal)
	} else {
		Message = Message + res
	}

	mountpoint, err := block.Mount(spc.Spec.Name)
	if err != nil {
		Message = Message + " unable to mount" + err.Error()
		//util.CheckErr(err, util.Fatal)
	} else {
		Message = " Mountpoint=" + mountpoint
		//util.CheckErr(err, util.Fatal)
	}

	sps := apis.StoragePoolSpec{Name: spc.Spec.Name, Format: spc.Spec.Format,
		Mountpoint: spc.Spec.Mountpoint,
		Nodename:   Nodename,
		Message:    Message,
	}
	sp := apis.StoragePool{Spec: sps}
	sp.Name = sps.Name

	return sp
}

func Createsp(sp apis.StoragePool, c *Controller, namespace string) {
	spCopy := sp.DeepCopy()
	spr, err := c.clientset.OpenebsV1alpha1().StoragePools(namespace).Create(spCopy)

	if err != nil {
		glog.Info("Unable to create sp", err)
	} else {
		glog.Info("Created sp :", spr.Spec.Name)
	}
}

func Updatesp(sp apis.StoragePool, c *Controller, namespace string) {
	spCopy := sp.DeepCopy()
	spGot, err := c.clientset.OpenebsV1alpha1().StoragePools(namespace).Get(spCopy.Spec.Name, metav1.GetOptions{})
	spGot.Spec = spCopy.Spec
	spr, err := c.clientset.OpenebsV1alpha1().StoragePools(namespace).Update(spGot)

	if err != nil {
		glog.Info("Unable to update sp", err)
	} else {
		glog.Info("Updated sp :", spr.Spec.Name)
	}
}

func Deletesp(name string, c *Controller, namespace string) {
	err := c.clientset.OpenebsV1alpha1().StoragePools(namespace).Delete(name, &metav1.DeleteOptions{})

	if err != nil {
		glog.Info("Unable to delete sp ", name, err)
	} else {
		glog.Info("Deleted sp :", name)
	}
}

func IsDiskAvailable(name string) bool {
	var resJsonDecoded v1.BlockDeviceInfo
	err := block.ListBlockExec(&resJsonDecoded)
	if err != nil {
		return false
	}
	for _, blk := range resJsonDecoded.Blockdevices {
		if blk.Name == name {
			return true
		}
	}
	return false
}
