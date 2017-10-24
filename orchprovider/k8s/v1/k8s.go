// This file registers Kubernetes as an orchestration provider plugin in maya
// api server. This orchestration is for persistent volume provisioners which
// also are registered in maya api server.
package k8s

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/maya/orchprovider"
	"github.com/openebs/maya/types/v1"
	volProfile "github.com/openebs/maya/volume/profiles"
	k8sCoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	k8sExtnsV1Beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	//k8sUnversioned "k8s.io/client-go/pkg/api/unversioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sApiV1 "k8s.io/client-go/pkg/api/v1"
	k8sApisExtnsBeta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// K8sOrchestrator is a concrete implementation of following
// interfaces:
//
//  1. orchprovider.OrchestratorInterface,
//  2. orchprovider.NetworkPlacements &
//  3. orchprovider.StoragePlacements
type k8sOrchestrator struct {
	// label specified to this orchestrator
	label v1.NameLabel

	// name of the orchestrator as registered in the registry
	name v1.OrchProviderRegistry

	// k8sUtlGtr provides the handle to fetch K8sUtilInterface
	// NOTE:
	//    This will be set at runtime.
	k8sUtlGtr K8sUtilGetter

	// rCount is the number of replica counts
	//rCount int

	// vsm is the name of the openebs volume
	//vsm string

	// rImg is the replica image
	//rImg string

	// persistPath is the persistent path i.e. storage backend of the openebs volume
	//persistPath string

	// pvc is an instance of persistent volume claim which is provided during
	// openebs volume operations
	//pvc *v1.PersistentVolumeClaim

	// k8sClient is an instance of K8sClient that exposes Kubernetes based
	// operations
	//k8sClient K8sClient

	// dOps is an instance capable of performing Kubernetes related Deployment
	// operations
	//dOps k8sExtnsV1Beta1.DeploymentInterface

	// nodeSelKeysByName contains a mapping of replica identifier & respective
	// node selector key
	//nodeSelKeysByName map[string]string

	// nodeSelOpsByName contains a mapping of replica identifier & respective
	// node selector operator
	//nodeSelOpsByName map[string]string

	// nodeSelValuesByName contains a mapping of replica identifier & respective
	// node selector value
	//nodeSelValuesByName map[string]string
}

// NewK8sOrchestrator provides a new instance of K8sOrchestrator.
func NewK8sOrchestrator(label v1.NameLabel, name v1.OrchProviderRegistry) (orchprovider.OrchestratorInterface, error) {

	glog.Infof("Building '%s':'%s' orchestration provider", label, name)

	if string(label) == "" {
		return nil, fmt.Errorf("Label not found while building k8s orchestrator")
	}

	if string(name) == "" {
		return nil, fmt.Errorf("Name not found while building k8s orchestrator")
	}

	return &k8sOrchestrator{
		label: label,
		name:  name,
	}, nil
}

// Label provides the label assigned against this orchestrator.
// This is an implementation of the orchprovider.OrchestratorInterface interface.
func (k *k8sOrchestrator) Label() string {
	// TODO
	// Do not typecast. Make it typed
	// Ensure this for all orch provider implementors
	return string(k.label)
}

// Name provides the name of this orchestrator.
// This is an implementation of the orchprovider.OrchestratorInterface interface.
func (k *k8sOrchestrator) Name() string {
	// TODO
	// Do not typecast. Make it typed
	// Ensure this for all orch provider implementors
	return string(k.name)
}

// TODO
// Deprecate in favour of orchestrator profile
// Region is not supported by k8sOrchestrator.
// This is an implementation of the orchprovider.OrchestratorInterface interface.
func (k *k8sOrchestrator) Region() string {
	return ""
}

//
//func (k *k8sOrchestrator) getReplicaCount(volProfile volProfile.VolumeProvisionerProfile) (int, error) {
// is already set ?
//	if k.rCount != 0 {
//		return k.rCount, nil
//	}
// else extract it
//	rCount, err := volProfile.ReplicaCount()
//	if err != nil {
//		return 0, err
//	}
// set it so that it can be used later as well
//k.rCount = rCount
//return k.rCount, nil
//}

//
//func (k *k8sOrchestrator) getVSM(volProfile volProfile.VolumeProvisionerProfile) (string, error) {
// is already set ?
//if k.vsm != "" {
//	return k.vsm, nil
//}
// else extract it
//vsm, err := volProfile.VSMName()
//if err != nil {
//	return "", err
//}
// set it so that it can be used later as well
//k.vsm = vsm
//return k.vsm, nil
//}

//
//func (k *k8sOrchestrator) getReplicaImg(volProfile volProfile.VolumeProvisionerProfile) (string, error) {
// is already set ?
//if k.rImg != "" {
//	return k.rImg, nil
//}
// else extract it
//rImg, err := volProfile.ReplicaImage()
//if err != nil {
//	return "", err
//}
// set it so that it can be used later as well
//k.rImg = rImg
//return k.rImg, nil
//}

//
//func (k *k8sOrchestrator) getPersistPath(volProfile volProfile.VolumeProvisionerProfile) (string, error) {
// is already set ?
//if k.persistPath != "" {
//	return k.persistPath, nil
//}
// else extract it
//persistPath, err := volProfile.PersistentPath()
//if err != nil {
//	return "", err
//}
// set it so that it can be used later as well
//k.persistPath = persistPath
//return k.persistPath, nil
//}

//
//func (k *k8sOrchestrator) getPVC(volProfile volProfile.VolumeProvisionerProfile) (*v1.PersistentVolumeClaim, error) {
// is already set ?
//if k.pvc != nil {
//	return k.pvc, nil
//}
// else extract it
//pvc, err := volProfile.PVC()
//if err != nil {
//	return nil, err
//}
// set it so that it can be used later as well
//k.pvc = pvc
//return k.pvc, nil
//}

//
//func (k *k8sOrchestrator) getK8sClient(volProfile volProfile.VolumeProvisionerProfile) (K8sClient, error) {
// is already set ?
//if k.k8sClient != nil {
//	return k.k8sClient, nil
//}
// else extract it
//k8sUtl := k8sOrchUtil(k, volProfile)
//k8sClient, supported := k8sUtl.K8sClient()
//if !supported {
//	return nil, fmt.Errorf("K8s client not supported by '%s'", k8sUtl.Name())
//}
// set it so that it can be used later as well
//k.k8sClient = k8sClient
//return k.k8sClient, nil
//}

//
//func (k *k8sOrchestrator) getDeploymentOps(volProfile volProfile.VolumeProvisionerProfile) (k8sExtnsV1Beta1.DeploymentInterface, error) {
// is already set ?
//if k.dOps != nil {
//	return k.dOps, nil
//}
// else extract it
//kc, err := k.getK8sClient(volProfile)
//if err != nil {
//	return nil, err
//}
// fetch k8s deployment operations
//dOps, err := kc.DeploymentOps()
//if err != nil {
//	return nil, err
//}
// set it so that it can be used later as well
//k.dOps = dOps
//return k.dOps, nil
//}

// getNodeSelectorKey gets the key to select a node for replica placement.
// repIdentifier can be the replica pod name or a user provided replica pod
// identifier to associate the node selector key with its corresponding node
// selector value.
//func (k *k8sOrchestrator) getNodeSelectorKey(repIdentifier string, volProfile volProfile.VolumeProvisionerProfile) (string, error) {
// is already set ?
//if k.nodeSelKeysByName != nil && k.nodeSelKeysByName[repIdentifier] != "" {
//	return k.nodeSelKeysByName[repIdentifier], nil
//}
// else extract it
//nodeSelKey := volProfile.NodeSelectorKey(repIdentifier)

// set it (blank value is valid as well) so that it can be used later as well
//k.nodeSelKeysByName[repIdentifier] = nodeSelKey
//return k.nodeSelKeysByName[repIdentifier], nil
//}

// getNodeSelectorOp gets the operator to select a node for replica placement.
// repIdentifier can be the replica pod name or a user provided replica pod
// identifier to associate the node selector operator with its corresponding node
// selector key & node selector value.
//func (k *k8sOrchestrator) getNodeSelectorOp(repIdentifier string, volProfile volProfile.VolumeProvisionerProfile) (string, error) {
// is already set ?
//if k.nodeSelOpsByName != nil && k.nodeSelOpsByName[repIdentifier] != "" {
//	return k.nodeSelOpsByName[repIdentifier], nil
//}
// else extract it
//nodeSelOp := volProfile.NodeSelectorOp(repIdentifier)

// set it (blank value is valid as well) so that it can be used later as well
//k.nodeSelOpsByName[repIdentifier] = nodeSelOp
//return k.nodeSelOpsByName[repIdentifier], nil
//}

// getNodeSelectorValues gets the value(s) to select a node for replica placement.
// repIdentifier can be the replica pod name or a user provided replica pod
// identifier to associate the node selector value with its corresponding node
// selector key.
//func (k *k8sOrchestrator) getNodeSelectorValues(repName string, volProfile volProfile.VolumeProvisionerProfile) (string, error) {
// is already set ?
//if k.nodeSelValuesByName != nil && k.nodeSelValuesByName[repName] != "" {
//	return k.nodeSelValuesByName[repName], nil
//}
// else extract it
//nodeSelVal := volProfile.NodeSelectorValue(repName)

// set it (blank value is valid as well) so that it can be used later as well
//k.nodeSelValuesByName[repName] = nodeSelVal
//return k.nodeSelValuesByName[repName], nil
//}

// TODO
// Check if StorageOps() can do these stuff in a better way. This method &
// k8sOrchUtil() were introduced to inject mock dependency while unit testing.
//
// GetK8sUtil provides the k8sUtil instance that is capable of performing low
// level k8s operations
//
// NOTE:
//    This is an implementation of K8sUtilGetter interface
//
// NOTE:
//    This contract implementation helps to provide a custom instance
// of K8sUtilInterface if required. K8sUtilInterface is a external dependency of
// k8sOrchestrator. This method enables a loosely coupled way to set dependency.
func (k *k8sOrchestrator) GetK8sUtil(volProfile volProfile.VolumeProvisionerProfile) K8sUtilInterface {
	return &k8sUtil{
		volProfile: volProfile,
	}
}

// TODO
// Check if StorageOps() can do these stuff in a better way. This method &
// GetK8sUtil() were introduced to inject mock dependency while unit testing.
//
// k8sOrchUtil provides a utility function for k8sOrchestrator to get an
// instance of k8sUtilInterface
func k8sOrchUtil(k *k8sOrchestrator, volProfile volProfile.VolumeProvisionerProfile) K8sUtilInterface {
	if k.k8sUtlGtr == nil {
		// This is possible as k8sOrchestrator is a k8sUtilGetter implementor
		k.k8sUtlGtr = k
	}

	return k.k8sUtlGtr.GetK8sUtil(volProfile)
}

// TODO
//    Will it be good if this accepts VolumeProvisionerProfile and updates the
// k8sOrchestrator instance's properties ?
//
// StorageOps provides storage operations instance that deals with all storage
// related functionality by aligning with Kubernetes as the orchestration provider.
//
// NOTE:
//    This is an implementation of the orchprovider.OrchestratorInterface interface.
//
// NOTE:
//    This is invoked on a per request basis. In other words, every request will
// invoke StorageOps to invoke storage specific operations thereafter.
func (k *k8sOrchestrator) StorageOps() (orchprovider.StorageOps, bool) {
	return k, true
}

// AddStorage will add persistent volume running as containers. In OpenEBS
// terms AddStorage will add a VSM.
func (k *k8sOrchestrator) AddStorage(volProProfile volProfile.VolumeProvisionerProfile) (*v1.Volume, error) {

	// TODO
	// This is jiva specific
	// Move this entire logic to a separate package that will couple jiva
	// provisioner with k8s orchestrator

	// create k8s service of persistent volume controller
	_, err := k.createControllerService(volProProfile)
	if err != nil {
		k.DeleteStorage(volProProfile)
		return nil, err
	}

	// Get the persistent volume controller service name & IP address
	_, clusterIP, err := k.getControllerServiceDetails(volProProfile)
	if err != nil {
		k.DeleteStorage(volProProfile)
		return nil, err
	}

	// create k8s pod of persistent volume controller
	_, err = k.createControllerDeployment(volProProfile, clusterIP)
	if err != nil {
		k.DeleteStorage(volProProfile)
		return nil, err
	}

	_, err = k.createReplicaDeployment(volProProfile, clusterIP)
	if err != nil {
		k.DeleteStorage(volProProfile)
		return nil, err
	}

	// TODO
	// This is a temporary type that is used
	// Will move to VSM type
	pv := &v1.Volume{}
	vsm, _ := volProProfile.VSMName()
	pv.Name = vsm

	return pv, nil
}

// DeleteStorage will remove the VSM. The logic is built in such a way that
// ensures genuinely repeated attempts do not get errored out.
//
// NOTE:
//    Current logic is an attempt to delete the dependents as cascading
// delete option is not available.
//
// NOTE:
//    This also handles the cases where creation failed mid-flight, and bail
// out requires calling delete function.
func (k *k8sOrchestrator) DeleteStorage(volProProfile volProfile.VolumeProvisionerProfile) (bool, error) {
	// Assume the presence of atleast one VSM object
	// Set this flag to false initially
	var hasAtleastOneVSMObj bool

	if volProProfile == nil {
		return false, fmt.Errorf("Nil volume provisioner profile provided")
	}

	vsm, err := volProProfile.VSMName()
	if err != nil {
		return false, err
	}

	if strings.TrimSpace(vsm) == "" {
		return false, fmt.Errorf("VSM name is required to delete storage")
	}

	k8sUtl := k8sOrchUtil(k, volProProfile)

	kc, supported := k8sUtl.K8sClient()
	if !supported {
		return false, fmt.Errorf("K8s client not supported by '%s'", k8sUtl.Name())
	}

	// fetch k8s deployment operations
	dOps, err := kc.DeploymentOps()
	if err != nil {
		return false, err
	}

	// fetch k8s Pod operations
	pOps, err := kc.Pods()
	if err != nil {
		return false, err
	}

	// fetch k8s service operations
	sOps, err := kc.Services()
	if err != nil {
		return false, err
	}

	// This ensures the dependents of Deployment e.g. ReplicaSets to be deleted
	orphanDependents := false

	// Delete the Replica Deployments first
	rDeploys, err := k.getReplicaDeploys(vsm, dOps)
	if err != nil {
		return false, err
	}

	if rDeploys != nil && len(rDeploys.Items) > 0 {
		hasAtleastOneVSMObj = true
		for _, rd := range rDeploys.Items {
			err = dOps.Delete(rd.Name, &metav1.DeleteOptions{
				OrphanDependents: &orphanDependents,
			})
			if err != nil {
				return false, err
			}
		}
	}

	// Delete the Controller Deployments next
	cDeploys, err := k.getControllerDeploys(vsm, dOps)
	if err != nil {
		return false, err
	}

	if cDeploys != nil && len(cDeploys.Items) > 0 {
		hasAtleastOneVSMObj = true
		for _, cd := range cDeploys.Items {
			err = dOps.Delete(cd.Name, &metav1.DeleteOptions{
				OrphanDependents: &orphanDependents,
			})
			if err != nil {
				return false, err
			}
		}
	}

	// Delete the Replica Pods before Controller Pod(s)
	rPods, err := k.getReplicaPods(vsm, pOps)
	if err != nil {
		return false, err
	}

	if rPods != nil && len(rPods.Items) > 0 {
		hasAtleastOneVSMObj = true
		for _, rPod := range rPods.Items {
			err = pOps.Delete(rPod.Name, &metav1.DeleteOptions{
				OrphanDependents: &orphanDependents,
			})
			if err != nil {
				return false, err
			}
		}
	}

	// Delete the Controller Pods next
	cPods, err := k.getControllerPods(vsm, pOps)
	if err != nil {
		return false, err
	}

	if cPods != nil && len(cPods.Items) > 0 {
		hasAtleastOneVSMObj = true
		for _, cPod := range cPods.Items {
			err = pOps.Delete(cPod.Name, &metav1.DeleteOptions{
				OrphanDependents: &orphanDependents,
			})
			if err != nil {
				return false, err
			}
		}
	}

	// Delete the Controller Services at last
	cSvcs, err := k.getControllerServices(vsm, sOps)
	if err != nil {
		return false, err
	}

	if cSvcs != nil && len(cSvcs.Items) > 0 {
		hasAtleastOneVSMObj = true
		for _, cSvc := range cSvcs.Items {
			err = sOps.Delete(cSvc.Name, &metav1.DeleteOptions{
				OrphanDependents: &orphanDependents,
			})
			if err != nil {
				return false, err
			}
		}
	}

	// Nothing to be deleted
	if !hasAtleastOneVSMObj {
		return false, nil
	}

	return true, nil
}

// ReadStorage will fetch information about the persistent volume
//func (k *k8sOrchestrator) ReadStorage(volProProfile volProfile.VolumeProvisionerProfile) (*v1.PersistentVolumeList, error) {
func (k *k8sOrchestrator) ReadStorage(volProProfile volProfile.VolumeProvisionerProfile) (*v1.Volume, error) {
	// volProProfile is expected to have the VSM name
	return k.readVSM("", volProProfile)
}

// readVSM will fetch information about a VSM
func (k *k8sOrchestrator) readVSM(vsm string, volProProfile volProfile.VolumeProvisionerProfile) (*v1.Volume, error) {

	// flag that checks if at-least one child object of VSM exists
	doesExist := false

	if volProProfile == nil {
		return nil, fmt.Errorf("Nil volume provisioner profile provided")
	}

	// fetch VSM from volume provisioner profile if not provided explicitly
	if vsm == "" {
		v, err := volProProfile.VSMName()
		if err != nil {
			return nil, err
		}
		vsm = v
	}

	k8sUtl := k8sOrchUtil(k, volProProfile)

	kc, supported := k8sUtl.K8sClient()
	if !supported {
		return nil, fmt.Errorf("K8s client not supported by '%s'", k8sUtl.Name())
	}

	// fetch k8s Deployment operator
	dOps, err := kc.DeploymentOps()
	if err != nil {
		return nil, err
	}

	// fetch k8s Service operator
	sOps, err := kc.Services()
	if err != nil {
		return nil, err
	}

	// fetch k8s Pod operator
	pOps, err := kc.Pods()
	if err != nil {
		return nil, err
	}

	ns, err := kc.NS()
	if err != nil {
		return nil, err
	}

	glog.Infof("Fetching info on VSM '%s: %s'", ns, vsm)

	annotations := map[string]string{}

	// Extract from Replica Deployments
	rDeploys, err := k.getReplicaDeploys(vsm, dOps)
	if err != nil {
		return nil, err
	}

	if rDeploys != nil && len(rDeploys.Items) > 0 {
		doesExist = true
		for _, rd := range rDeploys.Items {
			SetReplicaCount(rd, annotations)
			SetReplicaVolSize(rd, annotations)
		}
	} else {
		glog.Warningf("Missing Replica Deployment(s) for VSM '%s: %s'", ns, vsm)
	}

	// Extract from Controller Pods
	cPods, err := k.getControllerPods(vsm, pOps)
	if err != nil {
		return nil, err
	}

	if cPods != nil && len(cPods.Items) > 0 {
		doesExist = true
		for _, cp := range cPods.Items {
			SetControllerIPs(cp, annotations)
			SetControllerStatuses(cp, annotations)
		}
	} else {
		glog.Warningf("Missing Controller Pod(s) for VSM '%s: %s'", ns, vsm)
	}

	// Extract from Replica Pods
	rPods, err := k.getReplicaPods(vsm, pOps)
	if err != nil {
		return nil, err
	}

	if rPods != nil && len(rPods.Items) > 0 {
		doesExist = true
		for _, rp := range rPods.Items {
			SetReplicaIPs(rp, annotations)
			SetReplicaStatuses(rp, annotations)
		}
	} else {
		glog.Warningf("Missing Replica Pod(s) for VSM '%s: %s'", ns, vsm)
	}

	// Extract from Controller Services
	cSvcs, err := k.getControllerServices(vsm, sOps)
	if err != nil {
		return nil, err
	}

	if cSvcs != nil && len(cSvcs.Items) > 0 {
		doesExist = true
		for _, cSvc := range cSvcs.Items {
			SetISCSITargetPortals(cSvc, annotations)
			SetServiceStatuses(cSvc, annotations)
			SetControllerClusterIPs(cSvc, annotations)
		}
	} else {
		glog.Warningf("Missing Controller Service(s) for VSM '%s: %s'", ns, vsm)
	}

	if !doesExist {
		return nil, nil
	}

	SetIQN(vsm, annotations)

	// TODO
	// This is a temporary type that is used
	// Will move to VSM type
	pv := &v1.Volume{}
	pv.Name = vsm
	pv.Annotations = annotations

	glog.Infof("Info fetched successfully for VSM '%s: %s'", ns, vsm)

	return pv, nil
}

// ListStorage will list a collections of VSMs
func (k *k8sOrchestrator) ListStorage(volProProfile volProfile.VolumeProvisionerProfile) (*v1.VolumeList, error) {
	if volProProfile == nil {
		return nil, fmt.Errorf("Nil volume provisioner profile provided")
	}

	glog.Infof("Listing VSMs at orchestrator '%s: %s'", k.Label(), k.Name())

	dl, err := k.getVSMDeployments(volProProfile)
	if err != nil {
		return nil, err
	}

	if dl == nil || dl.Items == nil || len(dl.Items) == 0 {
		return nil, nil
	}

	pvl := &v1.VolumeList{}

	for _, d := range dl.Items {

		// consider either controller or replica to filter the VSMs
		// we are considering only controller
		if strings.Contains(d.Name, string(v1.ReplicaSuffix)) {
			continue
		}

		vsm := v1.SanitiseVSMName(d.Name)
		if vsm == "" {
			return nil, fmt.Errorf("VSM name could not be determined from K8s Deployment 'name: %s'", d.Name)
		}

		pv, err := k.readVSM(vsm, volProProfile)
		if err != nil {
			// Ignore the error of this particular VSM
			// Cases where this particular VSM might be in a creating or deleting state
			continue
		}
		pvl.Items = append(pvl.Items, *pv)
	}

	glog.Infof("Listed VSMs 'count: %d' at orchestrator '%s: %s'", len(pvl.Items), k.Label(), k.Name())

	return pvl, nil
}

// addNodeTolerationsToDeploy
func (k *k8sOrchestrator) addNodeTolerationsToDeploy(nodeTaintTolerations []string, deploy *k8sApisExtnsBeta1.Deployment) error {

	// nTT is expected to be in key=value:effect
	for _, nTT := range nodeTaintTolerations {
		kveArr := strings.Split(nTT, ":")
		if len(kveArr) != 2 {
			return fmt.Errorf("Invalid args '%s' provided for node taint toleration", nTT)
		}

		kv := kveArr[0]
		effect := strings.TrimSpace(kveArr[1])

		kvArr := strings.Split(kv, "=")
		if len(kvArr) != 2 {
			return fmt.Errorf("Invalid kv '%s' provided for node taint toleration", kv)
		}
		k := strings.TrimSpace(kvArr[0])
		v := strings.TrimSpace(kvArr[1])

		// Setting to blank to validate later
		e := k8sApiV1.TaintEffect("")

		// Supports only these two effects
		if string(k8sApiV1.TaintEffectNoExecute) == effect {
			e = k8sApiV1.TaintEffectNoExecute
		} else if string(k8sApiV1.TaintEffectNoSchedule) == effect {
			e = k8sApiV1.TaintEffectNoSchedule
		}

		if string(e) == "" {
			return fmt.Errorf("Invalid effect '%s' provided for node taint toleration", effect)
		}

		toleration := k8sApiV1.Toleration{
			Key:      k,
			Operator: k8sApiV1.TolerationOpEqual,
			Value:    v,
			Effect:   e,
		}

		tls := append(deploy.Spec.Template.Spec.Tolerations, toleration)
		deploy.Spec.Template.Spec.Tolerations = tls
	}

	return nil
}

// createControllerDeployment creates a persistent volume controller deployment in
// kubernetes
func (k *k8sOrchestrator) createControllerDeployment(volProProfile volProfile.VolumeProvisionerProfile, clusterIP string) (*k8sApisExtnsBeta1.Deployment, error) {
	// fetch VSM name
	vsm, err := volProProfile.VSMName()
	if err != nil {
		return nil, err
	}

	if clusterIP == "" {
		return nil, fmt.Errorf("VSM cluster IP is required to create controller for vsm 'name: %s'", vsm)
	}

	cImg, imgSupport, err := volProProfile.ControllerImage()
	if err != nil {
		return nil, err
	}

	if !imgSupport {
		return nil, fmt.Errorf("VSM '%s' requires a controller container image", vsm)
	}

	k8sUtl := k8sOrchUtil(k, volProProfile)

	kc, supported := k8sUtl.K8sClient()

	if !supported {
		return nil, fmt.Errorf("K8s client not supported by '%s'", k8sUtl.Name())
	}

	// fetch deployment operator
	dOps, err := kc.DeploymentOps()
	if err != nil {
		return nil, err
	}

	glog.Infof("Adding controller for VSM 'name: %s'", vsm)
	var tolerationSeconds int64 = 0

	deploy := &k8sApisExtnsBeta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: vsm + string(v1.ControllerSuffix),
			Labels: map[string]string{
				string(v1.VSMSelectorKey):               vsm,
				string(v1.VolumeProvisionerSelectorKey): string(v1.JivaVolumeProvisionerSelectorValue),
				string(v1.ControllerSelectorKey):        string(v1.JivaControllerSelectorValue),
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       string(v1.K8sKindDeployment),
			APIVersion: string(v1.K8sDeploymentVersion),
		},
		Spec: k8sApisExtnsBeta1.DeploymentSpec{
			Template: k8sApiV1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						string(v1.VSMSelectorKey):        vsm,
						string(v1.ControllerSelectorKey): string(v1.JivaControllerSelectorValue),
					},
				},
				Spec: k8sApiV1.PodSpec{
					// Ensure the controller gets EVICTED as soon as possible
					Tolerations: []k8sApiV1.Toleration{
						k8sApiV1.Toleration{
							Effect:            k8sApiV1.TaintEffectNoExecute,
							Key:               "node.alpha.kubernetes.io/notReady",
							Operator:          k8sApiV1.TolerationOpExists,
							TolerationSeconds: &tolerationSeconds,
						},
						k8sApiV1.Toleration{
							Effect:            k8sApiV1.TaintEffectNoExecute,
							Key:               "node.alpha.kubernetes.io/unreachable",
							Operator:          k8sApiV1.TolerationOpExists,
							TolerationSeconds: &tolerationSeconds,
						},
					},
					Containers: []k8sApiV1.Container{
						k8sApiV1.Container{
							Name:    vsm + string(v1.ControllerSuffix) + string(v1.ContainerSuffix),
							Image:   cImg,
							Command: v1.JivaCtrlCmd,
							Args:    v1.MakeOrDefJivaControllerArgs(vsm, clusterIP),
							Ports: []k8sApiV1.ContainerPort{
								k8sApiV1.ContainerPort{
									ContainerPort: v1.DefaultJivaISCSIPort(),
								},
								k8sApiV1.ContainerPort{
									ContainerPort: v1.DefaultJivaAPIPort(),
								},
							},
						},
					},
				},
			},
		},
	}

	// check if node level taint toleration is required ?
	nTTs, reqd, err := volProProfile.IsControllerNodeTaintTolerations()
	if err != nil {
		return nil, err
	}

	if reqd {
		err = k.addNodeTolerationsToDeploy(nTTs, deploy)
		if err != nil {
			return nil, err
		}
	}

	// add persistent volume controller deployment
	dd, err := dOps.Create(deploy)
	if err != nil {
		return nil, err
	}

	glog.Infof("Added controller 'name: %s'", deploy.Name)

	return dd, nil
}

// createReplicaDeployment creates one or more persistent volume deployment
// replica(s) in Kubernetes
func (k *k8sOrchestrator) createReplicaDeployment(volProProfile volProfile.VolumeProvisionerProfile, clusterIP string) (*k8sApisExtnsBeta1.Deployment, error) {
	// fetch VSM name
	vsm, err := volProProfile.VSMName()
	if err != nil {
		return nil, err
	}

	if clusterIP == "" {
		return nil, fmt.Errorf("VSM cluster IP is required to create replica(s) for vsm 'name: %s'", vsm)
	}

	rImg, err := volProProfile.ReplicaImage()
	if err != nil {
		return nil, err
	}

	rCount, err := volProProfile.ReplicaCount()
	if err != nil {
		return nil, err
	}

	//pCount, err := volProProfile.PersistentPathCount()
	//if err != nil {
	//	return nil, err
	//}

	//if pCount != rCount {
	//	return nil, fmt.Errorf("VSM '%s' replica count '%d' does not match persistent path count '%d'", vsm, rCount, pCount)
	//}

	pvc, err := volProProfile.PVC()
	if err != nil {
		return nil, err
	}

	// The position is always send as 1
	// We might want to get the replica index & send it
	// However, this does not matter if replicas are placed on different hosts !!
	//persistPath, err := volProProfile.PersistentPath(1, rCount)
	persistPath, err := volProProfile.PersistentPath()
	if err != nil {
		return nil, err
	}

	k8sUtl := k8sOrchUtil(k, volProProfile)

	kc, supported := k8sUtl.K8sClient()
	if !supported {
		return nil, fmt.Errorf("K8s client not supported by '%s'", k8sUtl.Name())
	}

	// fetch k8s deployment operator
	dOps, err := kc.DeploymentOps()
	if err != nil {
		return nil, err
	}

	// Create these many replicas -- if manual replica addition
	//for rcIndex := 1; rcIndex <= rCount; rcIndex++ {
	//glog.Infof("Adding replica #%d for VSM '%s'", rcIndex, vsm)

	glog.Infof("Adding replica(s) for VSM '%s'", vsm)

	deploy := &k8sApisExtnsBeta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			// -- if manual replica addition
			//Name: vsm + string(v1.ReplicaSuffix) + strconv.Itoa(rcIndex),
			Name: vsm + string(v1.ReplicaSuffix),
			Labels: map[string]string{
				string(v1.VSMSelectorKey):               vsm,
				string(v1.VolumeProvisionerSelectorKey): string(v1.JivaVolumeProvisionerSelectorValue),
				string(v1.ReplicaSelectorKey):           string(v1.JivaReplicaSelectorValue),
				// -- if manual replica addition
				//string(v1.ReplicaCountSelectorKey):      strconv.Itoa(rCount),
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       string(v1.K8sKindDeployment),
			APIVersion: string(v1.K8sDeploymentVersion),
		},
		Spec: k8sApisExtnsBeta1.DeploymentSpec{
			// -- automated K8s way of replica count management
			Replicas: rCount,
			Template: k8sApiV1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						string(v1.VSMSelectorKey):     vsm,
						string(v1.ReplicaSelectorKey): string(v1.JivaReplicaSelectorValue),
					},
				},
				Spec: k8sApiV1.PodSpec{
					// Ensure the replicas stick to its placement node even if the node dies
					// In other words DO NOT EVICT these replicas
					Tolerations: []k8sApiV1.Toleration{
						k8sApiV1.Toleration{
							Effect:   k8sApiV1.TaintEffectNoExecute,
							Key:      "node.alpha.kubernetes.io/notReady",
							Operator: k8sApiV1.TolerationOpExists,
						},
						k8sApiV1.Toleration{
							Effect:   k8sApiV1.TaintEffectNoExecute,
							Key:      "node.alpha.kubernetes.io/unreachable",
							Operator: k8sApiV1.TolerationOpExists,
						},
					},
					Affinity: &k8sApiV1.Affinity{
						// Inter-pod anti-affinity rule to spread the replicas across K8s minions
						PodAntiAffinity: &k8sApiV1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []k8sApiV1.PodAffinityTerm{
								k8sApiV1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											string(v1.VSMSelectorKey):     vsm,
											string(v1.ReplicaSelectorKey): string(v1.JivaReplicaSelectorValue),
										},
									},
									// TODO
									// This is host based inter-pod anti-affinity
									// Make it generic s.t. it can be zone based or region based
									// inter-pod anti-affinity as well.
									//
									// TODO
									// How about the cases, where some replicas should be host
									// based anti-affinity & other replicas should be zone based
									// anti-affinity. However, storage Admin should not spend effort
									// on this. There should be some intelligent mechanism which
									// can understand the setup to check if it has access to different
									// zones, regions, etc. In addition, this intelligence should
									// take into account storage capable nodes in these zones,
									// regions. All of these sould result in suggestions to maya
									// api server during provisioning.
									//
									// TODO
									// Considering above scenarios, it might make more sense to have
									// separate K8s Deployment for each replica. However,
									// there are dis-advantages in diverging from K8s replica set.
									TopologyKey: v1.GetPVPReplicaTopologyKey(pvc.Labels),
								},
							},
						},
					},
					Containers: []k8sApiV1.Container{
						k8sApiV1.Container{
							// -- if manual replica addition
							//Name:    vsm + string(v1.ReplicaSuffix) + string(v1.ContainerSuffix) + strconv.Itoa(rcIndex),
							Name:    vsm + string(v1.ReplicaSuffix) + string(v1.ContainerSuffix),
							Image:   rImg,
							Command: v1.JivaReplicaCmd,
							Args:    v1.MakeOrDefJivaReplicaArgs(pvc.Labels, clusterIP),
							Ports: []k8sApiV1.ContainerPort{
								k8sApiV1.ContainerPort{
									ContainerPort: v1.DefaultJivaReplicaPort1(),
								},
								k8sApiV1.ContainerPort{
									ContainerPort: v1.DefaultJivaReplicaPort2(),
								},
								k8sApiV1.ContainerPort{
									ContainerPort: v1.DefaultJivaReplicaPort3(),
								},
							},
							VolumeMounts: []k8sApiV1.VolumeMount{
								k8sApiV1.VolumeMount{
									Name:      v1.DefaultJivaMountName(),
									MountPath: v1.DefaultJivaMountPath(),
								},
							},
						},
					},
					Volumes: []k8sApiV1.Volume{
						k8sApiV1.Volume{
							Name: v1.DefaultJivaMountName(),
							VolumeSource: k8sApiV1.VolumeSource{
								HostPath: &k8sApiV1.HostPathVolumeSource{
									Path: persistPath,
								},
							},
						},
					},
				},
			},
		},
	}

	// check if node level taint toleration is required ?
	nTTs, reqd, err := volProProfile.IsReplicaNodeTaintTolerations()
	if err != nil {
		return nil, err
	}

	if reqd {
		err = k.addNodeTolerationsToDeploy(nTTs, deploy)
		if err != nil {
			return nil, err
		}
	}

	d, err := dOps.Create(deploy)
	if err != nil {
		return nil, err
	}

	glog.Infof("Successfully added replica(s) 'count: %d' for VSM '%s'", rCount, d.Name)

	//glog.Infof("Successfully added replica #%d for VSM '%s'", rcIndex, d.Name)
	//} -- end of for loop -- if manual replica addition

	return d, nil
}

// createControllerService creates a persistent volume controller service in
// kubernetes
func (k *k8sOrchestrator) createControllerService(volProProfile volProfile.VolumeProvisionerProfile) (*k8sApiV1.Service, error) {
	// fetch VSM name
	vsm, err := volProProfile.VSMName()
	if err != nil {
		return nil, err
	}

	k8sUtl := k8sOrchUtil(k, volProProfile)

	kc, supported := k8sUtl.K8sClient()
	if !supported {
		return nil, fmt.Errorf("K8s client not supported by '%s'", k8sUtl.Name())
	}

	// fetch k8s clientset & namespace
	sOps, err := kc.Services()
	if err != nil {
		return nil, err
	}

	// TODO
	// log levels & logging context to be taken care of
	glog.Infof("Adding service for Volume 'name : %s'", vsm)

	// TODO
	// Code this like a golang struct template
	// create persistent volume controller service
	svc := &k8sApiV1.Service{}
	svc.Kind = string(v1.K8sKindService)
	svc.APIVersion = string(v1.K8sServiceVersion)
	svc.Name = vsm + string(v1.ControllerSuffix) + string(v1.ServiceSuffix)
	svc.Labels = map[string]string{
		string(v1.VSMSelectorKey):               vsm,
		string(v1.VolumeProvisionerSelectorKey): string(v1.JivaVolumeProvisionerSelectorValue),
		string(v1.ServiceSelectorKey):           string(v1.JivaServiceSelectorValue),
	}

	iscsiPort := k8sApiV1.ServicePort{}
	iscsiPort.Name = string(v1.PortNameISCSI)
	iscsiPort.Port = v1.DefaultJivaISCSIPort()

	apiPort := k8sApiV1.ServicePort{}
	apiPort.Name = string(v1.PortNameAPI)
	apiPort.Port = v1.DefaultJivaAPIPort()

	svcSpec := k8sApiV1.ServiceSpec{}
	svcSpec.Ports = []k8sApiV1.ServicePort{iscsiPort, apiPort}
	// Set the selector that identifies the controller VSM
	svcSpec.Selector = map[string]string{
		string(v1.VSMSelectorKey):        vsm,
		string(v1.ControllerSelectorKey): string(v1.JivaControllerSelectorValue),
	}

	// Set the service spec
	svc.Spec = svcSpec

	// add controller service
	ssvc, err := sOps.Create(svc)

	// TODO
	// log levels & logging context to be taken care of
	if err == nil {
		glog.Infof("Added service 'name: %s'", svc.Name)
	}

	return ssvc, err
}

// getControllerServiceDetails fetches the service name & service IP address
// associated with the VSM
func (k *k8sOrchestrator) getControllerServiceDetails(volProProfile volProfile.VolumeProvisionerProfile) (string, string, error) {
	vsm, err := volProProfile.VSMName()
	if err != nil {
		return "", "", err
	}

	k8sUtl := k8sOrchUtil(k, volProProfile)

	kc, supported := k8sUtl.K8sClient()
	if !supported {
		return "", "", fmt.Errorf("K8s client is not supported by '%s'", k8sUtl.Name())
	}

	// fetch k8s service operations
	sOps, err := kc.Services()
	if err != nil {
		return "", "", err
	}

	svc, err := sOps.Get(vsm+string(v1.ControllerSuffix)+string(v1.ServiceSuffix), metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}

	return svc.Name, svc.Spec.ClusterIP, nil
}

// deleteService deletes the service associated with the provided VSM
func (k *k8sOrchestrator) deleteService(name string, volProProfile volProfile.VolumeProvisionerProfile) error {
	if name == "" {
		return fmt.Errorf("Name is required to delete the K8s Service")
	}

	k8sUtl := k8sOrchUtil(k, volProProfile)

	kc, supported := k8sUtl.K8sClient()
	if !supported {
		return fmt.Errorf("K8s client is not supported by '%s'", k8sUtl.Name())
	}

	// fetch k8s service operations
	sOps, err := kc.Services()
	if err != nil {
		return err
	}

	return sOps.Delete(name, &metav1.DeleteOptions{})
}

// getControllerServices fetches the Controller Services
func (k *k8sOrchestrator) getControllerServices(vsm string, serviceOps k8sCoreV1.ServiceInterface) (*k8sApiV1.ServiceList, error) {
	// filter the VSM Controller Services(s)
	lOpts := metav1.ListOptions{
		// A list of comma separated key=value filters will filter the
		// VSM Controller Service(s)
		LabelSelector: string(v1.VSMSelectorKeyEquals) + vsm + "," + string(v1.ServiceSelectorKeyEquals) + string(v1.JivaServiceSelectorValue),
	}

	sl, err := serviceOps.List(lOpts)
	if err != nil {
		return nil, err
	}

	return sl, nil
}

// getControllerDeploys fetches the Controller Deployments
func (k *k8sOrchestrator) getControllerDeploys(vsm string, deployOps k8sExtnsV1Beta1.DeploymentInterface) (*k8sApisExtnsBeta1.DeploymentList, error) {
	// filter the VSM Controller Deployment(s)
	lOpts := metav1.ListOptions{
		// A list of comma separated key=value filters will filter the
		// VSM Controller Deployment(s)
		LabelSelector: string(v1.VSMSelectorKeyEquals) + vsm + "," + string(v1.ControllerSelectorKeyEquals) + string(v1.JivaControllerSelectorValue),
	}

	dl, err := deployOps.List(lOpts)
	if err != nil {
		return nil, err
	}

	return dl, nil
}

// getReplicaDeploys fetches the Replica Deployments
func (k *k8sOrchestrator) getReplicaDeploys(vsm string, deployOps k8sExtnsV1Beta1.DeploymentInterface) (*k8sApisExtnsBeta1.DeploymentList, error) {
	// filter the VSM Replica Deployment(s)
	lOpts := metav1.ListOptions{
		// A list of comma separated key=value filters will filter the
		// VSM Replica Deployment(s)
		LabelSelector: string(v1.VSMSelectorKeyEquals) + vsm + "," + string(v1.ReplicaSelectorKeyEquals) + string(v1.JivaReplicaSelectorValue),
	}

	dl, err := deployOps.List(lOpts)
	if err != nil {
		return nil, err
	}

	return dl, nil
}

// getControllerPods fetches the Controller Pods
func (k *k8sOrchestrator) getControllerPods(vsm string, podOps k8sCoreV1.PodInterface) (*k8sApiV1.PodList, error) {
	// filter the VSM Controller Pod(s)
	pOpts := metav1.ListOptions{
		// A list of comma separated key=value filters will filter the
		// VSM Controller Pod(s)
		LabelSelector: string(v1.VSMSelectorKeyEquals) + vsm + "," + string(v1.ControllerSelectorKeyEquals) + string(v1.JivaControllerSelectorValue),
	}

	cp, err := podOps.List(pOpts)
	if err != nil {
		return nil, err
	}

	return cp, nil
}

// getReplicaPods fetches the Replica Pods
func (k *k8sOrchestrator) getReplicaPods(vsm string, podOps k8sCoreV1.PodInterface) (*k8sApiV1.PodList, error) {
	// filter the VSM Replica Pod(s)
	pOpts := metav1.ListOptions{
		// A list of comma separated key=value filters will filter the
		// VSM Replica Pod(s)
		LabelSelector: string(v1.VSMSelectorKeyEquals) + vsm + "," + string(v1.ReplicaSelectorKeyEquals) + string(v1.JivaReplicaSelectorValue),
	}

	rp, err := podOps.List(pOpts)
	if err != nil {
		return nil, err
	}

	return rp, nil
}

// getPods deletes the Pods w.r.t the VSM
func (k *k8sOrchestrator) getPods(vsm string, volProProfile volProfile.VolumeProvisionerProfile) (*k8sApiV1.PodList, error) {

	if strings.TrimSpace(vsm) == "" {
		return nil, fmt.Errorf("VSM name is required to get Pods")
	}

	k8sUtl := k8sOrchUtil(k, volProProfile)

	kc, supported := k8sUtl.K8sClient()
	if !supported {
		return nil, fmt.Errorf("K8s client not supported by '%s'", k8sUtl.Name())
	}

	// fetch k8s Pod operations
	pOps, err := kc.Pods()
	if err != nil {
		return nil, err
	}

	rps, err := k.getReplicaPods(vsm, pOps)
	if err != nil {
		return nil, err
	}

	cps, err := k.getControllerPods(vsm, pOps)
	if err != nil {
		return nil, err
	}

	// Merge the Replica & Controller Pods
	for _, rPod := range rps.Items {
		cps.Items = append(cps.Items, rPod)
	}

	return cps, nil
}

// getDeployment fetches the Deployment associated with the provided name of
// deployment
func (k *k8sOrchestrator) getDeployment(deployName string, volProProfile volProfile.VolumeProvisionerProfile) (*k8sApisExtnsBeta1.Deployment, error) {
	if strings.TrimSpace(deployName) == "" {
		return nil, fmt.Errorf("Deployment name is required to get its details")
	}

	k8sUtl := k8sOrchUtil(k, volProProfile)

	kc, supported := k8sUtl.K8sClient()
	if !supported {
		return nil, fmt.Errorf("K8s client not supported by '%s'", k8sUtl.Name())
	}

	// fetch k8s deployment operations
	dOps, err := kc.DeploymentOps()
	if err != nil {
		return nil, err
	}

	return dOps.Get(deployName, metav1.GetOptions{})
}

// getDeploymentList fetches the deployments associated with the provided VSM name
func (k *k8sOrchestrator) getDeploymentList(vsm string, volProProfile volProfile.VolumeProvisionerProfile) (*k8sApisExtnsBeta1.DeploymentList, error) {
	// fetch VSM if not provided
	if vsm == "" {
		v, err := volProProfile.VSMName()
		if err != nil {
			return nil, err
		}
		vsm = v
	}

	k8sUtl := k8sOrchUtil(k, volProProfile)

	kc, supported := k8sUtl.K8sClient()
	if !supported {
		return nil, fmt.Errorf("K8s client not supported by '%s'", k8sUtl.Name())
	}

	ns, err := kc.NS()
	if err != nil {
		return nil, err
	}

	dOps, err := kc.DeploymentOps()
	if err != nil {
		return nil, err
	}

	lOpts := metav1.ListOptions{
		LabelSelector: string(v1.VSMSelectorKeyEquals) + vsm,
	}

	deployList, err := dOps.List(lOpts)
	if err != nil {
		return nil, err
	}

	if deployList == nil {
		return nil, fmt.Errorf("VSM(s) '%s:%s' not found at orchestrator '%s:%s'", ns, vsm, k.Label(), k.Name())
	}

	return deployList, nil
}

// getVSMDeployments fetches all the VSM deployments
func (k *k8sOrchestrator) getVSMDeployments(volProProfile volProfile.VolumeProvisionerProfile) (*k8sApisExtnsBeta1.DeploymentList, error) {

	k8sUtl := k8sOrchUtil(k, volProProfile)

	kc, supported := k8sUtl.K8sClient()
	if !supported {
		return nil, fmt.Errorf("K8s client not supported by '%s'", k8sUtl.Name())
	}

	dOps, err := kc.DeploymentOps()
	if err != nil {
		return nil, err
	}

	// Filter the VSM deployments only
	// Filter it via the volume provisioner selector key as the name of the VSM is
	// unknown
	lOpts := metav1.ListOptions{
		LabelSelector: string(v1.VolumeProvisionerSelectorKey) + string(v1.SelectorEquals) + string(v1.JivaVolumeProvisionerSelectorValue),
	}

	vsmList, err := dOps.List(lOpts)
	if err != nil {
		return nil, err
	}

	if vsmList == nil || vsmList.Items == nil || len(vsmList.Items) == 0 {
		return nil, nil
	}

	return vsmList, nil
}

// getVSMServices fetches all the VSM services
func (k *k8sOrchestrator) getVSMServices(volProProfile volProfile.VolumeProvisionerProfile) (*k8sApiV1.ServiceList, error) {

	k8sUtl := k8sOrchUtil(k, volProProfile)

	kc, supported := k8sUtl.K8sClient()
	if !supported {
		return nil, fmt.Errorf("K8s client not supported by '%s'", k8sUtl.Name())
	}

	sOps, err := kc.Services()
	if err != nil {
		return nil, err
	}

	// Filter the VSM services only
	// Filter it via the volume provisioner selector key as the name of the VSM is
	// unknown
	lOpts := metav1.ListOptions{
		LabelSelector: string(v1.VolumeProvisionerSelectorKey) + string(v1.SelectorEquals) + string(v1.JivaVolumeProvisionerSelectorValue),
	}

	sList, err := sOps.List(lOpts)
	if err != nil {
		return nil, err
	}

	if sList == nil || sList.Items == nil || len(sList.Items) == 0 {
		return nil, nil
	}

	return sList, nil
}
