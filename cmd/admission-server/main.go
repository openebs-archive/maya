package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"github.com/openebs/maya/pkg/webhook"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	snapclientset "github.com/openebs/maya/pkg/client/generated/openebs.io/snapshot/v1alpha1/clientset/internalclientset"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	var parameters webhook.Parameters

	// get command line parameters
	flag.IntVar(&parameters.Port, "port", 443, "Webhook server port.")
	flag.StringVar(&parameters.CertFile, "tlsCertFile", "/etc/webhook/certs/cert.pem", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&parameters.KeyFile, "tlsKeyFile", "/etc/webhook/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")
	flag.Parse()

	// Get in cluster config
	cfg, err := getClusterConfig(kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}
	// Building Kubernetes Clientset
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}
	// Building OpenEBS Clientset
	openebsClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building openebs clientset: %s", err.Error())
	}

	// Building Snapshot Clientset
	snapClient, err := snapclientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building openebs snapshot clientset: %s", err.Error())
	}

	wh, err := webhook.New(parameters, kubeClient, openebsClient, snapClient)
	if err != nil {
		glog.Fatalf("failed to create validation webhook: %s", err.Error())
	}

	// define http server and server handler
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", wh.Serve)
	wh.Server.Handler = mux

	// start webhook server in new routine
	go func() {
		if err := wh.Server.ListenAndServeTLS("", ""); err != nil {
			glog.Errorf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	glog.Info("Webhook server started")

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-signalChan

	glog.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")
	err = wh.Server.Shutdown(context.Background())
	if err != nil {
		glog.Errorf("failed to shutdown server: error {%v}", err)
	}
}

// GetClusterConfig return the config for k8s.
func getClusterConfig(kubeconfig string) (*rest.Config, error) {
	var masterURL string
	cfg, err := rest.InClusterConfig()
	if err != nil {
		glog.Errorf("Failed to get k8s Incluster config. %+v", err)
		if kubeconfig == "" {
			return nil, fmt.Errorf("Kubeconfig is empty: %v", err.Error())
		}
		cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("Error building kubeconfig: %s", err.Error())
		}
	}
	return cfg, err
}
