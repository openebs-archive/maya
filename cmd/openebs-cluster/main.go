/*
Copyright 2019 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"os"

	log "github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io"
	mlogs "github.com/openebs/maya/pkg/logs"
	operator "github.com/openebs/maya/pkg/operator/openebscluster/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func main() {
	var metricsAddr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.Parse()
	mlogs.InitLogs()
	defer mlogs.FlushLogs()

	// Get a config to talk to the kubernetes apiserver
	log.Info("setting kubernetes client config")
	cfg, err := config.GetConfig()
	if err != nil {
		log.Errorf("failed to set up client config: %#v", err)
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	log.Info("setting up openebs cluster")
	mgr, err := manager.New(cfg, manager.Options{MetricsBindAddress: metricsAddr})
	if err != nil {
		log.Errorf("failed to set up openebs cluster: %#v", err)
		os.Exit(1)
	}

	log.Info("setting up scheme for all resources")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Errorf("failed to add APIs to scheme: %#v", err)
		os.Exit(1)
	}

	log.Info("setting up openebs cluster controllers")
	if err := operator.AddControllersToKubeManager(mgr); err != nil {
		log.Errorf("failed to register controllers to openebs cluster: %#v", err)
		os.Exit(1)
	}

	log.Info("starting openebs cluster")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Errorf("failed to start openebs cluster: %#v", err)
		os.Exit(1)
	}
}
