/*
Copyright 2018 The OpenEBS Authors

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

package template

import (
	"strings"
	"text/template"

	. "github.com/openebs/maya/pkg/task/v1alpha1"
	"github.com/openebs/maya/pkg/util"
)

// delete returns a new instance of delete based runtask command
//
// Examples:
// ---------
// {{- delete jiva volume | run -}}
// {{- delete cstor volume | run -}}
func delete(ml ...RunCommandMiddleware) *RunCommand {
	return RunCommandMiddlewareList(ml).Update(Command().DeleteAction())
}

// get returns a new instance of get based runtask command
//
// Examples:
// ---------
// {{- get jiva volume | run -}}
// {{- get cstor volume | run -}}
func get(ml ...RunCommandMiddleware) *RunCommand {
	return RunCommandMiddlewareList(ml).Update(Command().GetAction())
}

// list returns a new instance of list based runtask command
//
// Examples:
// ---------
// {{- list jiva volume | run -}}
// {{- list cstor volume | run -}}
func list(ml ...RunCommandMiddleware) *RunCommand {
	return RunCommandMiddlewareList(ml).Update(Command().ListAction())
}

// create returns a new instance of create based runtask command
//
// Examples:
// ---------
// {{- create jiva volume | run -}}
// {{- create cstor volume | run -}}
// {{- create cstor snapshot | run -}}
func create(ml ...RunCommandMiddleware) *RunCommand {
	return RunCommandMiddlewareList(ml).Update(Command().CreateAction())
}

// patch returns a new instance of patch based runtask command
//
// Examples:
// ---------
// {{- patch jiva volume | run -}}
// {{- patch cstor volume | run -}}
func patch(ml ...RunCommandMiddleware) *RunCommand {
	return RunCommandMiddlewareList(ml).Update(Command().PatchAction())
}

// update returns a new instance of update based runtask command
//
// Examples:
// ---------
// {{- update jiva volume | run -}}
// {{- update cstor volume | run -}}
func update(ml ...RunCommandMiddleware) *RunCommand {
	return RunCommandMiddlewareList(ml).Update(Command().UpdateAction())
}

// jiva returns a new instance of jiva based runtask command
//
// Examples:
// ---------
// {{- get jiva volume | run -}}
// {{- list jiva volume | run -}}
func jiva() RunCommandMiddleware {
	return JivaCategory()
}

// cstor returns a new instance of cstor based runtask command
//
// Examples:
// ---------
// {{- get cstor volume | run -}}
// {{- list cstor volume | run -}}
func cstor() RunCommandMiddleware {
	return CstorCategory()
}

// volume returns a new instance of volume based runtask command
//
// Examples:
// ---------
// {{- get cstor volume | run -}}
// {{- list cstor volume | run -}}
// {{- update jiva volume | run -}}
// {{- create jiva volume | run -}}
func volume() RunCommandMiddleware {
	return VolumeCategory()
}

// snapshot returns a new instance of snapshot based runtask command
//
// Examples:
// {{- create cstor snapshot | run -}}
func snapshot() RunCommandMiddleware {
	return SnapshotCategory()
}

// slect returns the values of the specified paths post runtask command
// execution
//
// Examples:
// ---------
// {{- select "all" | get cstor volume | run -}}
// {{- select "name" | list cstor volume | run -}}
// {{- select "name" "namespace" | update jiva volume | run -}}
// {{- select ".metadata.name" | create jiva volume | run -}}
func slect(paths ...string) RunCommandMiddleware {
	if len(paths) == 0 {
		paths = append(paths, "all")
	}
	return Select(paths)
}

// withoption sets the provided <key,value> pair as an input data to runtask command
//
// Examples:
// ---------
// {{- $url := "http://10.10.10.10:9501" -}}
// {{- delete jiva volume | withoption "url" $url | withoption "name" "myvol" | run -}}
func withoption(key, value string, given *RunCommand) (updated *RunCommand) {
	return WithData(given, key, value)
}

// run executes the runtask command
//
// Examples:
// ---------
// {{- get jiva volume | run -}}
func run(given *RunCommand) RunCommandResult {
	return given.Run()
}

// runlog executes the runtask command and stores the result as well as other
// information at the provided paths
//
// Examples:
// ---------
// {{- get jiva volume | runlog "getj.result" "getj.extras" .Values -}}
//
// NOTES:
// -----------
// '.Values' is of type map[string]interface{} & is provided to the
// template before template execution
//
// Once above template gets executed '.Values' will have .Values.getj.result
// and .Values.getj.extras set with result and extras due to runtask command
// execution
func runlog(resultpath, debugpath string, store map[string]interface{}, given *RunCommand) (res RunCommandResult) {
	res = run(given)

	resultpath = strings.TrimPrefix(resultpath, ".")
	debugpath = strings.TrimPrefix(debugpath, ".")

	// result store
	resultlocator := strings.Split(resultpath, ".")
	util.SetNestedField(store, res.Result(), resultlocator...)

	// debug store
	debuglocator := strings.Split(debugpath, ".")
	util.SetNestedField(store, res.Debug(), debuglocator...)
	return
}

// runCommandFuncs returns the set of runtask command based template functions
func runCommandFuncs() template.FuncMap {
	return template.FuncMap{
		"delete":     delete,
		"get":        get,
		"lst":        list,
		"create":     create,
		"patch":      patch,
		"update":     update,
		"jiva":       jiva,
		"cstor":      cstor,
		"volume":     volume,
		"snapshot":   snapshot,
		"withoption": withoption,
		"run":        run,
		"runlog":     runlog,
		"select":     slect,
	}
}
