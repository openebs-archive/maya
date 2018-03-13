package main

import (
	"os"

	"github.com/openebs/maya/cmd/cstor-repl-ctrl/app/command"
	cstorlogger "github.com/openebs/maya/pkg/logs"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// Run cstor-repl-ctrl
func run() error {
	// Init logging
	cstorlogger.InitLogs()
	defer cstorlogger.FlushLogs()

	// Create & execute new command
	cmd, err := command.NewCstorReplCtrl()
	if err != nil {
		return err
	}

	return cmd.Execute()
}
