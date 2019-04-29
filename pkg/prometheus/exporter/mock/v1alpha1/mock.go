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

package v1alpha1

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// BuildRegex returns the regular expression look alike prometheus's
// response to get request.
func BuildRegex(list []string) []*regexp.Regexp {
	regexList := make([]*regexp.Regexp, 0)
	for _, r := range list {
		regexList = append(regexList, regexp.MustCompile(r))
	}
	return regexList
}

// PrometheusService is used to mock the behaviour of prometheus's client-go
// apis for testing purpose.
// Here, it registers the collector and then starts a test http server and send
// a get request to which server reply with the response as expected
func PrometheusService(col prometheus.Collector, stop chan struct{}) []byte {
	if err := prometheus.Register(col); err != nil {
		s := fmt.Sprintf("collector failed to register: %s", err)
		panic(s)
	}
	var server *httptest.Server
	start := make(chan struct{})
	go func(start, stop chan struct{}) {
		server = httptest.NewServer(promhttp.Handler())
		start <- struct{}{}
		<-stop
	}(start, stop)

	<-start
	client := http.DefaultClient
	client.Timeout = 5 * time.Second
	resp, err := client.Get(server.URL)
	if err != nil {
		s := fmt.Sprintf("unexpected failed response from prometheus: %s", err)
		panic(s)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s := fmt.Sprintf("failed reading server response: %s", err)
		panic(s)
	}
	return buf
}

// Unregister unregister the collector
func Unregister(col prometheus.Collector) bool {
	return prometheus.Unregister(col)
}
