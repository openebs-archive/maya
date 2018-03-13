package crdops

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
)

const poolControllerName = "cstorPool"

// CstorPoolController is the controller implementation for cstorPool resources.
type CstorPoolController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// clientset is a CRD package generated for custom API group.
	clientset clientset.Interface

	// cstorPoolSynced is used for caches sync to get populated
	cstorPoolSynced cache.InformerSynced

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

// NewCstorPoolController returns a new instance of controller
func NewCstorPoolController(
	kubeclientset kubernetes.Interface,
	clientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	cstorInformerFactory informers.SharedInformerFactory) *CstorPoolController {

	// obtain references to shared index informers for the cstorPool resources
	cstorPoolInformer := cstorInformerFactory.Openebs().V1alpha1().CstorPools()

	crdscheme.AddToScheme(scheme.Scheme)

	// Create event broadcaster to receive events and send them to any EventSink, watcher, or log.
	// Add NewCstorPoolController types to the default Kubernetes Scheme so Events can be
	// logged for CstorPool Controller types.
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)

	// StartEventWatcher starts sending events received from this EventBroadcaster to the given
	// event handler function. The return value can be ignored or used to stop recording, if
	// desired. Events("") denotes empty namespace
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: poolControllerName})

	controller := &CstorPoolController{
		kubeclientset:   kubeclientset,
		clientset:       clientset,
		cstorPoolSynced: cstorPoolInformer.Informer().HasSynced,
		workqueue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "cstorcrds"),
		recorder:        recorder,
	}

	glog.Info("Setting up event handlers")

	// Instantiating QueueLoad before entering workqueue.
	q := QueueLoad{}

	// Set up an event handler for when CstorPool resources change.
	cstorPoolInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			q.operation = "add"
			controller.enqueueCstorPool(obj, q)
		},
		UpdateFunc: func(old, new interface{}) {
			newCstorPool := new.(*apis.CstorPool)
			oldCstorPool := old.(*apis.CstorPool)
			// Periodic resync will send update events for all known CstorPool.
			// Two different versions of the same CstorPool will always have different RVs.
			if newCstorPool.ResourceVersion == oldCstorPool.ResourceVersion {
				return
			}
			q.operation = "update"
			controller.enqueueCstorPool(new, q)
		},
		DeleteFunc: func(obj interface{}) {
			q.operation = "delete"
			controller.enqueueCstorPool(obj, q)
		},
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *CstorPoolController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting CstorPool controller")

	// Wait for the k8s caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.cstorPoolSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting cstorPool workers")
	// Launch worker to process cstorPool resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started cstorPool workers")
	<-stopCh
	glog.Info("Shutting down cstorPool workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *CstorPoolController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *CstorPoolController) processNextWorkItem() bool {
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
		// cstorPool resource to be synced.
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
// converge the two. It then updates the Status block of the cstorPoolUpdated resource
// with the current status of the resource.
func (c *CstorPoolController) syncHandler(key, operation string) error {
	// Convert the key(namespace/name) string into a distinct name
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	cstorPoolUpdated, err := c.clientset.OpenebsV1alpha1().CstorPools().Get(name, metav1.GetOptions{})
	if err != nil {
		// The cstorPool resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("cstorPoolUpdated '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	switch operation {
	case "add":
		glog.Info("added event")

		err := checkValidPool(cstorPoolUpdated)
		if err != nil {
			return err
		}

		err = importPool(cstorPoolUpdated)
		if err == nil {
			return nil
		}

		err = createPool(cstorPoolUpdated)
		if err != nil {
			return err
		}
		break

	case "update":
		glog.Info("updated event")
		break

	case "delete":
		glog.Info("deleted event")
		break
	}

	return nil
}

// enqueueCstorPool takes a CstorPool resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than CstorPools.
func (c *CstorPoolController) enqueueCstorPool(obj interface{}, q QueueLoad) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	q.key = key
	c.workqueue.AddRateLimited(q)
}

// importPool imports cstor pool if already present.
func importPool(cstorPoolUpdated *apis.CstorPool) error {
	// populate pool import attributes.
	var importAttr []string
	importAttr = append(importAttr, "import")
	if cstorPoolUpdated.Spec.Poolspec.Cachefile != "" {
		cachefile := "cachefile=" + cstorPoolUpdated.Spec.Poolspec.Cachefile
		importAttr = append(importAttr, "-c", cachefile)
	}

	importAttr = append(importAttr, cstorPoolUpdated.Spec.Poolspec.Poolname)

	// execute import pool command.
	cmdimport := exec.Command("zpool", importAttr...)
	stdoutStderrImport, err := cmdimport.CombinedOutput()
	if err != nil {
		glog.Error("Pool import err: ", err)
		glog.Error("stdoutStderr: ", string(stdoutStderrImport))
		return err
	}

	glog.Info("Importing Pool Successful")
	return nil
}

// createPool creates a new cstor pool.
func createPool(cstorPoolUpdated *apis.CstorPool) error {
	// populate pool creation attributes.
	var createAttr []string
	createAttr = append(createAttr, "create", "-f", "-o")
	if cstorPoolUpdated.Spec.Poolspec.Cachefile != "" {
		cachefile := "cachefile=" + cstorPoolUpdated.Spec.Poolspec.Cachefile
		createAttr = append(createAttr, cachefile)
	}

	createAttr = append(createAttr, cstorPoolUpdated.Spec.Poolspec.Poolname)
	if len(cstorPoolUpdated.Spec.Disks) < 1 {
		return fmt.Errorf("Disk name(s) cannot be empty")
	}
	for _, disk := range cstorPoolUpdated.Spec.Disks {
		createAttr = append(createAttr, disk)
	}

	//execute pool creation command.
	poolCreateCmd := exec.Command("zpool", createAttr...)

	if glog.V(4) {
		glog.Info("poolCreateCmd : ", poolCreateCmd)
	}
	stdoutStderr, err := poolCreateCmd.CombinedOutput()
	if err != nil {
		glog.Error("err: ", err)
		glog.Error("stdoutStderr: ", string(stdoutStderr))
		return err
	}
	glog.Info("Creating Pool Successful")
	return nil
}

// checkValidPool checks for validity of cstor pool resource.
func checkValidPool(cstorPoolUpdated *apis.CstorPool) error {
	if cstorPoolUpdated.Spec.Poolspec.Poolname == "" {
		return fmt.Errorf("Poolname cannot be empty")
	}
	return nil
}
