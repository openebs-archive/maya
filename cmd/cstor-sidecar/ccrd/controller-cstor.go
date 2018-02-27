package ccrd

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"

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

const controllerAgentName = "cstor-sidecar"

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

	cstorLister listers.CstorCrdLister

	cstorSynced cache.InformerSynced

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
	cstorInformerFactory informers.SharedInformerFactory) *Controller {

	// obtain references to shared index informers for the sp and spc
	// types.
	cstorInformer := cstorInformerFactory.Openebs().V1alpha1().CstorCrds()

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
		cstorLister:   cstorInformer.Lister(),
		cstorSynced:   cstorInformer.Informer().HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "cstorcrds"),
		recorder:      recorder,
	}

	glog.Info("Setting up event handlers")
	// Set up an event handler for when spc resources change
	q := QueueLoad{}
	cstorInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			q.operation = "add"
			controller.enqueueCstorCrd(obj, q)
		},
		UpdateFunc: func(old, new interface{}) {
			newspc := new.(*apis.CstorCrd)
			oldspc := old.(*apis.CstorCrd)
			if newspc.ResourceVersion == oldspc.ResourceVersion {
				// Periodic resync will send update events for all known spc.
				// Two different versions of the same spc will always have different RVs.
				return
			}
			q.operation = "update"
			controller.enqueueCstorCrd(new, q)
		},
		DeleteFunc: func(obj interface{}) {
			q.operation = "delete"
			controller.enqueueCstorCrd(obj, q)
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
	glog.Info("Starting CstorCrd controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.cstorSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting cstor workers")
	// Launch two workers to process spc resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started cstor workers")
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
	_ = namespace
	cstorCrdUpdated, err := c.clientset.OpenebsV1alpha1().CstorCrds(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		// The spc resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("cstorCrdUpdated '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	switch operation {
	case "add":
		fmt.Println("added event")
		cachefile := "cachefile=" + cstorCrdUpdated.Spec.Zpool.Cachefile

		cmdimport := exec.Command("zpool", "import", "-c",
			cstorCrdUpdated.Spec.Zpool.Cachefile,
			cstorCrdUpdated.Spec.Zpool.Poolname)
		stdoutStderrImport, err := cmdimport.CombinedOutput()
		if err != nil {
			fmt.Println("err: ", err)
			fmt.Println("stdoutStderr: ", string(stdoutStderrImport))
		}else {
			fmt.Println("Importing Successful")
			return nil
		}

		cmd1 := exec.Command("zpool", "create", "-f", "-o", cachefile,
			cstorCrdUpdated.Spec.Zpool.Poolname,
			cstorCrdUpdated.Spec.Zpool.DiskPath)
		fmt.Println("cmd1 : ", cmd1)
		//res, err := cmd1.Output()
		stdoutStderr, err := cmd1.CombinedOutput()
		if err != nil {
			fmt.Println("err: ", err)
			fmt.Println("stdoutStderr: ", string(stdoutStderr))
			return err
		} else {
			var compression string
			var Readonly string
			logbias := "logbias=" + cstorCrdUpdated.Spec.Zpool.Zfs.Logbias
			copies := "copies=" + strconv.Itoa(cstorCrdUpdated.Spec.Zpool.Zfs.Copies)
			sync := "sync=" + cstorCrdUpdated.Spec.Zpool.Zfs.Sync
			fullvolname := cstorCrdUpdated.Spec.Zpool.Poolname + "/" + cstorCrdUpdated.Spec.Zpool.Zfs.Volname
			fmt.Println("stdoutStderr: ", string(stdoutStderr))
			if cstorCrdUpdated.Spec.Zpool.Zfs.Compression == true {
				compression = "compression=on"
			} else {
				compression = "compression=off"
			}
			if cstorCrdUpdated.Spec.Zpool.Zfs.Readonly == true {
				Readonly = "readonly=on"
			} else {
				Readonly = "readonly=off"
			}
			cmd2 := exec.Command("zfs", "create", "-s", "-b",
				cstorCrdUpdated.Spec.Zpool.Zfs.Blocksize,
				"-o", compression, "-o", logbias, "-o", copies,
				"-o", sync, "-o", Readonly,
				"-V", cstorCrdUpdated.Spec.Zpool.Zfs.Size, fullvolname)
			fmt.Println("cmd2 : ", cmd2)
			stdoutStderr, err := cmd2.CombinedOutput()
			if err != nil {
				fmt.Println("err: ", err)
				fmt.Println("stdoutStderr: ", string(stdoutStderr))
				return err
			} else {
				fmt.Println("stdoutStderr: ", string(stdoutStderr))
			}
		}
		break
	case "update":
		fmt.Println("updated event")
		break
	case "delete":
		fmt.Println("deleted event")
		break
	}

	return nil
}

// enqueueSpc takes a spc resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than spc.
func (c *Controller) enqueueCstorCrd(obj interface{}, q QueueLoad) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	q.key = key
	c.workqueue.AddRateLimited(q)
}
