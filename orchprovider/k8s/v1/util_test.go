package v1

import (
	"errors"
	"fmt"
	"testing"

	"github.com/openebs/maya/types/v1"
	volProfile "github.com/openebs/maya/volume/profiles"
	"k8s.io/client-go/kubernetes"
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
	pvc := &v1.Volume{
		Namespace: "default",
	}
	//pvc.Labels = map[string]string{}

	volP, _ := volProfile.GetDefaultVolProProfile(pvc)

	k8sUtl := &k8sUtil{
		volProfile: volP,
	}

	for i, c := range cases {

		incActual, err := k8sUtl.IsInCluster()
		if err != nil {
			t.Errorf("TestCase: '%d' ExpectedInClusterErr: 'nil' ActualInClusterErr: '%s'", i, err.Error())
		}

		if incActual != c.incluster {
			t.Errorf("TestCase: '%d' ExpectedInCluster: '%t' ActualInCluster: '%t'", i, c.incluster, incActual)
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
	pvc := &v1.Volume{
		Namespace: "ok",
	}
	//pvc.Labels = map[string]string{}

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
	pvc := &v1.Volume{
		Namespace: "ok",
	}
	//pvc.Labels = map[string]string{}

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

// TestNS tests the namespace property of k8sUtil instance.
func TestNS(t *testing.T) {

	cases := []struct {
		ns string
	}{
		{"default"},
		{"test"},
	}

	for i, c := range cases {

		// a noop pvc that in turn signals use of defaults
		pvc := &v1.Volume{
			Namespace: c.ns,
		}

		volP, _ := volProfile.GetDefaultVolProProfile(pvc)

		k8sUtl := &k8sUtil{
			volProfile: volP,
		}

		nsActual, err := k8sUtl.NS()
		if err != nil {
			t.Errorf("TestCase: '%d' ExpectedNSErr: 'nil' ActualNSErr: '%s'", i, err.Error())
		}

		if nsActual != c.ns {
			t.Errorf("TestCase: '%d' ExpectedNS: '%s' ActualNS: '%s'", i, c.ns, nsActual)
		}
	}
}

// TestgetOutCluster tests TestgetOutClusterCS func
func TestGetOutClusterCS(t *testing.T) {

	cases := []struct {
		name           string
		expectedOutput *kubernetes.Clientset
		expectedError  error
	}{
		{"default", nil, errors.New("out cluster clientset not supported in 'k8sutil @ 'default''")},
		{"test", nil, errors.New("out cluster clientset not supported in 'k8sutil @ 'test''")},
	}

	for i, val := range cases {
		pvc := &v1.Volume{
			Namespace: val.name,
		}

		volP, _ := volProfile.GetDefaultVolProProfile(pvc)

		k8sUtl := &k8sUtil{
			volProfile: volP,
		}

		out, err := k8sUtl.getOutClusterCS()

		if out != val.expectedOutput {
			t.Errorf("TestCase: '%d' Expected Output :%v but got :%v", i, val.expectedOutput, out)
		}
		if err.Error() != val.expectedError.Error() {
			t.Errorf("TestCase: '%d' Expected Error :%v but got :%v", i, val.expectedError, err)
		}
	}
}

// TestIsInCluster tests the output of IsIncluster func
func TestIsInCluster(t *testing.T) {

	cases := []struct {
		name           string
		expectedOutput bool
		expectedError  error
	}{
		{"default", true, nil},
		{"test", true, nil},
	}

	for i, val := range cases {
		pvc := &v1.Volume{
			Namespace: val.name,
		}

		volP, _ := volProfile.GetDefaultVolProProfile(pvc)

		k8sUtl := &k8sUtil{
			volProfile: volP,
		}

		out, err := k8sUtl.IsInCluster()

		if out != val.expectedOutput {
			t.Errorf("TestCase: '%d' ExpectedOutput %v but got :%v", i, val.expectedOutput, out)
		}
		if err != val.expectedError {
			t.Errorf("TestCase: '%d' ExpectedError %v but got :%v", i, val.expectedError, err)
		}
	}
}
