package main

import (
	"os"

	"github.com/mitchellh/cli"
	"github.com/openebs/maya/command"
)

// Commands returns the mapping of CLI commands for Maya. The meta
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
		"setup-omm": func() (cli.Command, error) {
			return &command.InstallMayaCommand{
				M: meta,
			}, nil
		},
		"setup-osh": func() (cli.Command, error) {
			return &command.InstallOpenEBSCommand{
				M: meta,
			}, nil
		},
		"omm-status": func() (cli.Command, error) {
			return &command.ServerMembersCommand{
				Meta: meta,
			}, nil
		},
		"osh-status": func() (cli.Command, error) {
			return &command.NodeStatusCommand{
				Meta: meta,
			}, nil
		},

		"vsm-create": func() (cli.Command, error) {
			return &command.VsmCreateCommand{
				M: meta,
			}, nil
		},
		"vsm-list": func() (cli.Command, error) {
			return &command.VsmListCommand{
				M: meta,
			}, nil
		},
		"vsm-update": func() (cli.Command, error) {
			return &command.VsmUpdateCommand{
				M: meta,
			}, nil
		},
		"vsm-stop": func() (cli.Command, error) {
			return &command.VsmStopCommand{
				Meta: meta,
			}, nil
		},
		"network-install": func() (cli.Command, error) {
			return &command.NetworkInstallCommand{
				M: meta,
			}, nil
		},
		"version": func() (cli.Command, error) {
			ver := Version
			rel := VersionPrerelease
			if GitDescribe != "" {
				ver = GitDescribe
				// Trim off a leading 'v', we append it anyways.
				if ver[0] == 'v' {
					ver = ver[1:]
				}
			}
			if GitDescribe == "" && rel == "" && VersionPrerelease != "" {
				rel = "dev"
			}

			return &command.VersionCommand{
				Revision:          GitCommit,
				Version:           ver,
				VersionPrerelease: rel,
				Ui:                meta.Ui,
			}, nil
		},
	}
}
