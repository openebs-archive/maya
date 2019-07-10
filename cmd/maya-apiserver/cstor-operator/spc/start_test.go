/*
Copyright 2017 The OpenEBS Authors

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
package spc

import (
	"sync"
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	var err error
	var errchannel = make(chan error)
	go func() {
		var mux = sync.RWMutex{}
		err = Start(&mux)
		errchannel <- err
	}()
	select {
	case err1 := <-errchannel:
		err = err1
	case <-time.After(5 * time.Second):
		err = nil
	}
	if err == nil {
		t.Fatal("Error should not be nil as no incluster config is present")
	}
}
