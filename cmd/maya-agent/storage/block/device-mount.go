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

package block

import (
	"fmt"
	"os/exec"
)

//Mount is to mount the specified disk to /mnt/<disk>
func Mount(disk string) {
	mountpoint := "/mnt/" + disk
	diskDev := "/dev/" + disk
	//p flag is to return no error if the directory is available already
	res, err := exec.Command("mkdir", "-p", mountpoint).Output()
	if err != nil {
		panic(err)
	}

	res, err = exec.Command("mount", diskDev, mountpoint).Output()
	if err != nil {
		panic(err)
	}
	if len(res) == 0 {
		fmt.Println("Successfully mounted on: ", mountpoint)
	} else {
		fmt.Println("Mounting failure")
	}
}
