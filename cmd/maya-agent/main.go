package main

import (
	"github.com/openebs/maya/cmd/maya-agent/app"
	mayalogger "github.com/openebs/maya/kit/logs"
	"os"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// Run maya-agent
func run() error {
	// Init logging
	mayalogger.InitLogs()
	defer mayalogger.FlushLogs()

	// Create & execute new command
	cmd, err := app.NewMayaAgent()
	if err != nil {
		return err
	}

	return cmd.Execute()
}
