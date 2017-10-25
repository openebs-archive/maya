package main

import (
	"os"

	"github.com/mitchellh/cli"
	"github.com/openebs/maya/cmd/maya-apiserver/app/command"
	"github.com/openebs/maya/pkg/version"
)

// Commands returns the mapping of CLI commands for Maya server. The meta
// parameter lets you set meta options for all commands.
func Commands(metaPtr *command.Meta) map[string]cli.CommandFactory {
	if metaPtr == nil {
		metaPtr = new(command.Meta)
	}

	meta := *metaPtr
	if meta.Ui == nil {
		meta.Ui = &cli.BasicUi{
			Reader:      os.Stdin,
			Writer:      os.Stdout,
			ErrorWriter: os.Stderr,
		}
	}

	return map[string]cli.CommandFactory{
		"up": func() (cli.Command, error) {
			return &command.UpCommand{
				Revision:          version.GetGitCommit(),
				Version:           version.GetVersion(),
				VersionPrerelease: version.GetBuildMeta(),
				Ui:                meta.Ui,
				ShutdownCh:        make(chan struct{}),
			}, nil
		},
		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Revision:          version.GetGitCommit(),
				Version:           version.GetVersion(),
				VersionPrerelease: version.GetBuildMeta(),
				Ui:                meta.Ui,
			}, nil
		},
	}
}
