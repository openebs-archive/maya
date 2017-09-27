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
		//the following CLI commands are deprecated with latest implementation.
		//In kubernetes environment, it is no longer required to setup
		//openebs master and host.
		/*
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
		*/
		"volume": func() (cli.Command, error) {
			return &command.VolumeCommand{}, nil
		},
		"volume create": func() (cli.Command, error) {
			return &command.VsmCreateCommand{
				Meta: meta,
			}, nil
		},
		"volume list": func() (cli.Command, error) {
			return &command.VsmListCommand{
				Meta: meta,
			}, nil
		},
		"volume stats": func() (cli.Command, error) {
			return &command.VsmStatsCommand{
				Meta: meta,
			}, nil
		},
		/*	"volume update": func() (cli.Command, error) {
			return &command.VsmUpdateCommand{
				M: meta,
			}, nil
		},*/
		"volume delete": func() (cli.Command, error) {
			return &command.VsmStopCommand{
				Meta: meta,
			}, nil
		},
		"snapshot": func() (cli.Command, error) {
			return &command.SnapshotCommand{}, nil
		},

		"snapshot create": func() (cli.Command, error) {
			return &command.SnapshotCreateCommand{
				Meta: meta,
			}, nil
		},
		"snapshot list": func() (cli.Command, error) {
			return &command.SnapshotListCommand{
				Meta: meta,
			}, nil
		},

		/*	"snapshot rm": func() (cli.Command, error) {
			return &command.SnapshotDeleteCommand{
				Meta: meta,
			}, nil
		},*/
		"snapshot revert": func() (cli.Command, error) {
			return &command.SnapshotRevertCommand{
				Meta: meta,
			}, nil
		},

		/*	"network-install": func() (cli.Command, error) {
				return &command.NetworkInstallCommand{
					M: meta,
				}, nil
			},
		*/
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
