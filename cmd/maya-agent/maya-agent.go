package main

import "os"
import "github.com/openebs/maya/cmd/maya-agent/app"

//import k8slogsutil "k8s.io/kubernetes/pkg/util/logs"

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// Run maya-agent
func run() error {
	// Init logging
	//k8slogsutil.InitLogs()
	//defer k8slogsutil.FlushLogs()

	// Create & execute new command
	cmd, err := app.NewMayaAgent()
	if err != nil {
		return err
	}

	return cmd.Execute()
}
