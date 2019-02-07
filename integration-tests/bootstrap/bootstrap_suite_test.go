package bootstrap

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeConfigPath *string
	kubeConfig     *kubernetes.Clientset
)

func homeDir() (string, error) {
	if h := os.Getenv("HOME"); h != "" { // linux
		return h, nil
	} else if h := os.Getenv("USERPROFILE"); h != "" { // windows
		return h, nil
	}
	return "", fmt.Errorf("Not able to locate home directory")
}

var _ = BeforeSuite(func() {
	home, err := homeDir()
	Expect(err).To(BeNil())

	// Parse the kube config path
	kubeConfigPath = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfigPath)
	Expect(err).To(BeNil())

	// create the clientset
	kubeConfig, err = kubernetes.NewForConfig(config)
	Expect(err).To(BeNil())
})

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bootstrap")
}
