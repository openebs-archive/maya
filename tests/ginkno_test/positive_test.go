/*
Copyright 2019 The OpenEBS Authors

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

package ginkgo

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	"github.com/openebs/maya/tests/cstor"
)

var _ = Describe("positive volume", func() {
	BeforeEach(func() {
		By(fmt.Sprintf("BEFORE EACH POSITIVE TEST CASE"))
	})
	When("Positive tests", func() {
		It("positive tests", func() {
			By(fmt.Sprintf("POSITIVE TEST CASE IT doesn't run positive test case\n"))
			By(fmt.Sprintf("Value is %s", cstor.KubeConfigPath))
		})
	})
	AfterEach(func() {
		By(fmt.Sprintf("AFTER EACH POSITIVE TEST CASE"))
	})
})
