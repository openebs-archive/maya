/*
Copyright 2018 The OpenEBS Authors.

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

package controller

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/maya/cmd/cstor-pool-mgmt/cstorops/pool"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a resource is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a resource fails
	// to sync due to a resource of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events which
	// fails to sync due to a resource already existing
	MessageResourceExists = "Resource %q already exists and cannot be handled"
	// MessageResourceSynced is the message used for an Event fired when a resource
	// is synced successfully
	MessageResourceSynced = "Resource synced successfully"
)

// QueueLoad is for storing the key and type of operation before entering workqueue
type QueueLoad struct {
	key       string
	operation string
}

// PoolNameHandler tries to get pool name, if error, then block forever or
// tries for particular number of attempts.
func PoolNameHandler(isBlockForever bool) (string, error) {
	for cnt := 0; ; cnt++ {
		poolname, err := pool.GetPoolName()
		if err != nil {
			glog.Infof("Attempt %v: Waiting for Pool..", cnt)
			time.Sleep(5 * time.Second)
			if isBlockForever {
				continue
			} else if cnt > 3 {
				return "", err
			}
		} else {
			return poolname, nil
		}
	}
}

func execShResult(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("Missing command")
	}

	out := &bytes.Buffer{}
	cmd := exec.Command("/bin/sh", "-c", s)
	cmd.Stdin = os.Stdin
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	r := string(out.Bytes())
	return r, nil
}
