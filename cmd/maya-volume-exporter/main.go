package main

import (
	"os"

	"github.com/openebs/maya/cmd/maya-volume-exporter/app/command"
	mayalogger "github.com/openebs/maya/pkg/logs"
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
	cmd, err := command.NewCmdVolumeExporter()
	if err != nil {
		return err
	}

	return cmd.Execute()
}
