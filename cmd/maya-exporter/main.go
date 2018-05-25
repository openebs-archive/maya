package main

import (
	"log"
	"os"

	"github.com/openebs/maya/cmd/maya-exporter/app/command"
	mayalogger "github.com/openebs/maya/pkg/logs"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// Run maya-exporter
func run() error {
	// Init logging
	mayalogger.InitLogs()
	defer mayalogger.FlushLogs()

	// Create & execute new command
	cmd, err := command.NewCmdVolumeExporter()
	if err != nil {
		log.Println("Can't execute the command, found err :", err)
		return err
	}

	return cmd.Execute()
}
