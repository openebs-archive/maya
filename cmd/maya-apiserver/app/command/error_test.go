/*
Copyright 2017 The OpenEBS Authors.

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
package command

import (
    "errors"
    "os"
    "os/exec"
    "testing"
)

func TestCheckError(t *testing.T) {
    if os.Getenv("CRASH") == "1" {
        err := errors.New("Some error")
        CheckError(err)
        return
    }
    cmd := exec.Command(os.Args[0], "-test.run=TestCheckError")
    cmd.Env = append(os.Environ(), "CRASH=1")
    err := cmd.Run()
    if e, ok := err.(*exec.ExitError); ok && !e.Success() {
        return
    }
    t.Fatalf("process ran with error %v, want exit status 1", err)
}
