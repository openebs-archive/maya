// Copyright © 2018-2019 The OpenEBS Authors
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

package mapiserver

import "github.com/openebs/maya/pkg/util"

const (
	getStatusPath = "/latest/meta-data/instance-id"
)

//TODO: check before mayactl decouple is done
// GetStatus returns the status of maya-apiserver via http
func GetStatus() (string, error) {
	body, err := getRequest(GetURL()+getStatusPath, "", false)
	if err != nil {
		return "Connection failed", err
	}
	if string(body) != `"any-compute"` {
		err = util.ErrServerUnavailable
	}
	return string(body), err
}
