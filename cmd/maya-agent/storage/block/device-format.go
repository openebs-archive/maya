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

//Format is to format the disk with specified type
func Format(disk, ftype string) {
	diskDev := "/dev/" + disk
	fmt.Println("diskDev:", diskDev)
	fmt.Println("ftype:", ftype)
	res, err := exec.Command("mkfs", "-F", "-t", ftype, diskDev).Output()
	if err != nil {
		panic(err)
	}
	fmt.Println("res:", string(res))

}
