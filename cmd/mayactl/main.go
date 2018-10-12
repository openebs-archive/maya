// Copyright © 2017 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"

	"github.com/openebs/maya/cmd/mayactl/app/command"
	"github.com/posener/complete"
)

func main() {
	snapshot := complete.Command{
		Sub: complete.Commands{
			"create": complete.Command{},
			"list": complete.Command{},
			"revert": complete.Command{},
		},
	}

	version := complete.Command{
		Flags: complete.Flags{
			"--help": complete.PredictNothing,
			"-h": complete.PredictNothing,
		},
	}

	volume := complete.Command{
		Sub: complete.Commands{
			"delete": complete.Command{},
			"info": complete.Command{},
			"list": complete.Command{},
			"stats": complete.Command{},
		},
	}

	help := complete.Command{
		Sub: complete.Commands{
			"help": complete.Command{},
			"snapshot": complete.Command{},
			"version": complete.Command{},
			"volume": complete.Command{},
		},
		Flags: complete.Flags{
			"--help": complete.PredictNothing,
			"-h": complete.PredictNothing,
		},
	}

	cmp := complete.New(
		"mayactl",
		complete.Command{
			Sub: complete.Commands{
				"help": help,
				"snapshot": snapshot,
				"version": version,
				"volume": volume,
			},
			GlobalFlags: complete.Flags{
				"--alsologtostderr": complete.PredictNothing,
				"--log_backtrace_at": complete.PredictAnything,
				"--log_dir": complete.PredictAnything,
				"--logtostderr": complete.PredictNothing,
				"--mapiserver": complete.PredictAnything,
				"-m": complete.PredictAnything,
				"--mapiserverport": complete.PredictAnything,
				"-p": complete.PredictAnything,
				"--namespace": complete.PredictAnything,
				"-n": complete.PredictAnything,
				"--stderrthreshold": complete.PredictAnything,
				"-v": complete.PredictAnything,
				"--v": complete.PredictAnything,
				"--vmodule": complete.PredictAnything,
			},
		},
	)

	cmp.CLI.InstallName = "complete"
	cmp.CLI.UninstallName = "uncomplete"
	cmp.AddFlags(nil)

	flag.Parse()

	if cmp.Complete() {
		return
	}

	err := command.NewMayaCommand().Execute()
	
	command.CheckError(err)
}
