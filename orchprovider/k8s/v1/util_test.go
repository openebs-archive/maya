package k8s

import (
	"fmt"
	"testing"

	"github.com/openebs/maya/types/v1"
	volProfile "github.com/openebs/maya/volume/profiles"
)

// TestK8sUtilInterfaceCompliance verifies if k8sUtil implements
// all the exposed methods of the desired interfaces.
//
// NOTE:
//    In case of non-compliance, this logic will error out during compile
// time itself.
func TestK8sUtilInterfaceCompliance(t *testing.T) {
	// k8sUtil implements K8sUtilInterface
	var _ K8sUtilInterface = &k8sUtil{}
	// k8sUtil implements K8sClients
	var _ K8sClient = &k8sUtil{}
}

// TestK8sUtil tests the k8sUil instance as well as its properties
func TestK8sUtil(t *testing.T) {
	cases := []struct {
		name      string
		incluster bool
		ns        string
	}{
		{"k8sutil", true, "default"},
	}

	// a noop pvc that in turn signals use of defaults
	pvc := &v1.Volume{}
	pvc.Labels = map[string]string{}

	volP, _ := volProfile.GetDefaultVolProProfile(pvc)

	k8sUtl := &k8sUtil{
		volProfile: volP,
	}

	for i, c := range cases {

		incActual, err := k8sUtl.InCluster()
		if err != nil {
			t.Errorf("TestCase: '%d' ExpectedInClusterErr: 'nil' ActualInClusterErr: '%s'", i, err.Error())
		}

		if incActual != c.incluster {
			t.Errorf("TestCase: '%d' ExpectedInCluster: '%s' ActualInCluster: '%s'", i, c.incluster, incActual)
		}

		nsActual, err := k8sUtl.NS()
		if err != nil {
			t.Errorf("TestCase: '%d' ExpectedNSErr: 'nil' ActualNSErr: '%s'", i, err.Error())
		}

		if nsActual != c.ns {
			t.Errorf("TestCase: '%d' ExpectedNS: '%s' ActualNS: '%s'", i, c.ns, nsActual)
		}

		nActual := k8sUtl.Name()
		nExptd := fmt.Sprintf("%s @ '%s'", c.name, c.ns)

		if nActual != nExptd {
			t.Errorf("TestCase: '%d' ExpectedName: '%s' ActualName: '%s'", i, nExptd, nActual)
		}
	}
}

// TestK8sUtilPods tests the working of Pods() method.
//
// NOTE:
//    Error is expected when this test is run on environment that does not run
// k8s.
func TestK8sUtilPods(t *testing.T) {
	cases := []struct {
		err string
	}{
		{"unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined"},
	}

	// a noop pvc that in turn signals use of defaults
	pvc := &v1.Volume{}
	pvc.Labels = map[string]string{}

	volP, _ := volProfile.GetDefaultVolProProfile(pvc)

	k8sUtl := &k8sUtil{
		volProfile: volP,
	}

	for i, c := range cases {
		_, err := k8sUtl.Pods()

		if err != nil && err.Error() != c.err {
			t.Errorf("TestCase: '%d' ExpectedPodsErr: '%s' ActualPodsErr: '%s'", i, c.err, err.Error())
		}
	}
}

// TestK8sUtilServices tests the working of Services() method.
//
// NOTE:
//    Error is expected when this test is run on environment that does not run
// k8s.
func TestK8sUtilServices(t *testing.T) {
	cases := []struct {
		err string
	}{
		{"unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined"},
	}

	// a noop pvc that in turn signals use of defaults
	pvc := &v1.Volume{}
	pvc.Labels = map[string]string{}

	volP, _ := volProfile.GetDefaultVolProProfile(pvc)

	k8sUtl := &k8sUtil{
		volProfile: volP,
	}

	for i, c := range cases {
		_, err := k8sUtl.Services()

		if err != nil && err.Error() != c.err {
			t.Errorf("TestCase: '%d' ExpectedServicesErr: '%s' ActualServicesErr: '%s'", i, c.err, err.Error())
		}
	}
}
