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

package v1alpha1

import (
	"fmt"
	. "github.com/openebs/maya/pkg/client/http/v1alpha1"
	jp "github.com/openebs/maya/pkg/jsonpath/v1alpha1"
)

// jivaVolumeDelete represents a jiva volume delete runtask command
//
// NOTE:
//  This is an implementation of CommandRunner
type jivaVolumeDelete struct {
	cmd *RunCommand
}

// Run deletes jiva volume contents
func (j *jivaVolumeDelete) Run() (r RunCommandResult) {
	// api call to list volumes and volume actions per controller
	baseurl, _ := j.cmd.Data["url"].(string)
	if len(baseurl) == 0 {
		return j.cmd.AddError(fmt.Errorf("missing base url: failed to delete jiva volume")).Result(nil)
	}
	b, err := API("GET", baseurl, "volumes")
	if err != nil {
		return j.cmd.AddError(err).Result(nil)
	}

	// api call to delete jiva volume data
	durl := j.fetchDeleteVolumeLink(b)
	if len(durl) == 0 {
		return j.cmd.AddError(fmt.Errorf("delete action link not found: failed to delete jiva volume")).Result(nil)
	}
	b, err = URL("DELETE", durl)
	if err != nil {
		return j.cmd.AddError(err).Result(nil)
	}
	return j.cmd.Result(b)
}

// fetchDeleteVolumeLink fetches the url to delete jiva volume contents
func (j *jivaVolumeDelete) fetchDeleteVolumeLink(b []byte) (url string) {
	if b == nil {
		j.cmd.AddError(fmt.Errorf("nil volume actions: failed to fetch jiva volume delete link"))
		return
	}

	// extract delete action link based on volume name
	volname, _ := j.cmd.Data["name"].(string)
	if len(volname) == 0 {
		j.cmd.AddError(fmt.Errorf("missing volume name: failed to fetch jiva volume delete link"))
		return
	}

	// build the json query path
	p := fmt.Sprintf("{.data[?(@.name=='%s')].actions.deletevolume}", volname)
	jpath := jp.JSONPath("delete-jiva-volume").WithTargetAsRaw(b)

	// execute json query
	ul := jpath.Query(jp.SelectionList{jp.Selection("dellink", p)})

	// collect the messages occured during jsonpath querying
	j.cmd.Msgs.Merge(jpath.Msgs)
	return ul.ValueByName("dellink")
}
