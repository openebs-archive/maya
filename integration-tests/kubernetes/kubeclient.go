package kubernetes

import (
	"fmt"
	"os"

	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeConfigPath *string
	clientset      *kube.Clientset
)

// HomeDir return the Home Directory of the environement
func HomeDir() (string, error) {
	if h := os.Getenv("HOME"); h != "" { // linux
		return h, nil
	} else if h := os.Getenv("USERPROFILE"); h != "" { // windows
		return h, nil
	}

	return "", fmt.Errorf("Not able to locate home directory")
}

// GetConfigPath returns the path of kubeconfig
func GetConfigPath() (kubeConfigPath string, err error) {
	home, err := HomeDir()
	if err != nil {
		return
	}

	// Parse the kube config path
	kubeConfigPath = home + "/.kube/config"
	return kubeConfigPath, err
}

// GetClientSet returns the clientset for interacting the kubernetes cluster
func GetClientSet() (cl *kube.Clientset, err error) {
	kubeConfigPath, err := GetConfigPath()
	if err != nil {
		return
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return
	}

	// create the clientset
	return kube.NewForConfig(config)
}
