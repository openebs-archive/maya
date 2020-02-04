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
)

var _ = Describe("negative volume", func() {
	BeforeEach(func() {
		By(fmt.Sprintf("BEFORE EACH NEGATIVE TEST CASE"))
	})
	When("Negative test", func() {
		It("Negative tests", func() {
			By(fmt.Sprintf("NEGATIVE TEST CASE IT doesn't run positive test case\n"))
		})
	})
	It("Negative tests", func() {
		By(fmt.Sprintf("NEGATIVE TEST CASE IT doesn't run positive test case\n"))
	})
	AfterEach(func() {
		By(fmt.Sprintf("AFTER EACH NEGATIVE TEST CASE"))
	})
})
