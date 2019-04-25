// Copyright © 2019 The OpenEBS Authors
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

package iscsi

import (
	"fmt"
	"os/exec"
)

//IscsiLogin is to login to block devices in Storage area network
func IscsiLogin(target string) {
	var res []byte
	var err error
	if target == "all" {
		res, err = exec.Command("iscsiadm", "-m", "node", "-l").Output()
	} else {
		res, err = exec.Command("iscsiadm", "-m", "node", "-p", target, "-l").Output()
	}
	if err != nil {
		panic(err)
	}
	fmt.Println("Device(s) :", string(res))
}
