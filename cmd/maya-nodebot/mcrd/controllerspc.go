package mcrd

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/openebs/maya/cmd/maya-nodebot/storage/block"

	"github.com/golang/glog"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
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

	spapis "github.com/openebs/maya/pkg/storagepool-apis/openebs.io/v1"
	spclientset "github.com/openebs/maya/pkg/storagepool-client/clientset/versioned"
	//storagepoolinformers "github.com/openebs/maya/crd-code-generation/pkg/storagepool-client/informers/externalversions"

	spcapisv1alpha1 "github.com/openebs/maya/pkg/storagepoolclaim-apis/openebs.io/v1"
	spcclientset "github.com/openebs/maya/pkg/storagepoolclaim-client/clientset/versioned"
	spcscheme "github.com/openebs/maya/pkg/storagepoolclaim-client/clientset/versioned/scheme"
	spcinformers "github.com/openebs/maya/pkg/storagepoolclaim-client/informers/externalversions"
	spclisters "github.com/openebs/maya/pkg/storagepoolclaim-client/listers/example/v1"
	//samplev1alpha1 "github.com/testsamplecontroller/sample-controller/temp"
)

const controllerAgentName = "maya-nodebot"

const (
	// SuccessSynced is used as part of the Event 'reason' when a spc is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a spc fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
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
	spcclientset spcclientset.Interface

	storagepoolclient spclientset.Interface

	spcLister spclisters.StoragepoolclaimLister

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

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	spcclientset spcclientset.Interface,
	storagepoolclient spclientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	spcInformerFactory spcinformers.SharedInformerFactory) *Controller {

	// obtain references to shared index informers for the Deployment and spc
	// types.
	spcInformer := spcInformerFactory.Example().V1().Storagepoolclaims()

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	spcscheme.AddToScheme(scheme.Scheme)

	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:     kubeclientset,
		spcclientset:      spcclientset,
		spcLister:         spcInformer.Lister(),
		spcSynced:         spcInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "spcs"),
		recorder:          recorder,
		storagepoolclient: storagepoolclient,
	}

	glog.Info("Setting up event handlers")
	// Set up an event handler for when spc resources change
	spcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueSpcAdd,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueSpcUpdate(new)
		},
		DeleteFunc: controller.enqueueSpcDelete,
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
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// spc resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
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
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the spcUpdated resource with this namespace/name
	spcUpdated, err := c.spcLister.Storagepoolclaims(namespace).Get(name)
	if err != nil {
		// The spc resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("spcUpdated '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}
	//fmt.Println("spcUpdated : ", spcUpdated)
	_ = spcUpdated
	return nil
}

func (c *Controller) updateSpcStatus(spc *spcapisv1alpha1.Storagepoolclaim, deployment *appsv1beta2.Deployment) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	spcCopy := spc.DeepCopy()
	//spcCopy.Status.AvailableReplicas = 1
	// Until #38113 is merged, we must use Update instead of UpdateStatus to
	// update the Status block of the spc resource. UpdateStatus will not
	// allow changes to the Spec of the resource, which is ideal for ensuring
	// nothing other than resource status has been updated.
	_, err := c.spcclientset.ExampleV1().Storagepoolclaims(spc.Namespace).Update(spcCopy)
	return err
}

// enqueueSpc takes a spc resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than spc.
func (c *Controller) enqueueSpcAdd(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)

	c.handleObjectAdd(obj)
}

func (c *Controller) enqueueSpcUpdate(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
	//printing since the update corner case is yet to be decided
	fmt.Println("key updated: ", key)
	//fmt.Println("obj updated: ", obj)
}

func (c *Controller) enqueueSpcDelete(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)

	//simply printing the deleted key
	fmt.Println("key deleted: ", key)
	//fmt.Println("obj deleted: ", obj)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the spc resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that spc resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObjectAdd(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		glog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	glog.V(4).Infof("Processing object: %s", object.GetName())

	spc, err := c.spcclientset.ExampleV1().
		Storagepoolclaims(object.GetNamespace()).
		Get(object.GetName(), metav1.GetOptions{})
	if err != nil {
		fmt.Println("error", err)
	}
	//fmt.Println("spc", spc)

	fmt.Println("ADDED : ", spc.Spec.Name, spc.Spec.Format, spc.Spec.Mountpoint)

	Message := ""

	Nodename, err := GetNodeName()
	if err != nil {
		Message = Message + "unable to get nodename "
	}

	res, err := block.Format(spc.Spec.Name, spc.Spec.Format)
	if err != nil {
		Message = Message + " unable to format"
		//util.CheckErr(err, util.Fatal)
	} else {
		Message = res
	}
	mountpoint, err := block.Mount(spc.Spec.Name)
	if err != nil {
		Message = Message + " unable to mount"
		//util.CheckErr(err, util.Fatal)
	} else {
		Message = " Mountpoint=" + mountpoint
	}
	sps := spapis.StoragepoolSpec{Name: spc.Spec.Name, Format: spc.Spec.Format,
		Mountpoint: spc.Spec.Mountpoint,
		Nodename:   Nodename,
		Message:    Message,
	}

	sp := spapis.Storagepool{Spec: sps}
	sp.GenerateName = sps.Name
	spr, err := c.storagepoolclient.ExampleV1().Storagepools(object.GetNamespace()).Create(&sp)

	if err != nil {
		fmt.Println("Unable to create sp", err)
	} else {
		fmt.Println("Created sp", spr.Spec)
	}
	Message = ""
}

func GetNodeName() (string, error) {
	var Nodename string
	NodenameByte, err := exec.Command("uname", "-n").Output()
	Nodename = string(NodenameByte)
	return Nodename, err
}
