# CITF

[![Build Status](https://travis-ci.org/openebs/CITF.svg?branch=master)](https://travis-ci.org/openebs/CITF)
[![Go Report](https://goreportcard.com/badge/github.com/openebs/CITF)](https://goreportcard.com/report/github.com/openebs/CITF)
[![codecov](https://codecov.io/gh/openebs/CITF/branch/master/graph/badge.svg)](https://codecov.io/gh/openebs/CITF)
[![GoDoc](https://godoc.org/github.com/openebs/CITF?status.svg)](https://godoc.org/github.com/openebs/CITF)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/openebs/CITF/blob/master/LICENSE)

**Common Integration Test Framework** is a framework that will be used organization wide for Integration Test of all OpenEBS projects.

This repository is intended to only expose generic function which will help developers in writing Integration Tests. Though it won't produce any deliverable alone.

## Directory Structure in the Project
```
OpenEBS/project
   ├── integration_test
   │   ├── project_specific_package_for_integration_test
   │   │   ├── ...
   │   │   └── files.go
   │   ├── scenario1_test.go
   │   ├── scenario2_test.go
   │   ├── ...
   │   └── scenarioN_test.go
   ├── project_specific_packages
   └── vendor
       ├── package_related_vendors
       ├── ...
       └── github.com/OpenEBS/CITF
```

> Note: Developer should keep `integration_test` completely decoupled from the rest of the project packages.

## Instantiation

Developer has to instantiate CITF using `citf.NewCITF` function, which will initialize it with all the configurations specified by `citfoptions.CreateOptions` passed to it. 

> You should not pass `K8sInclude` in `citfoptions.CreateOptions` if your environment is not yet set. otherwise it will through error.

> If you want all options except `K8sInclude` in `CreateOptions` to set to `true`; you may use `citfoptions.CreateOptionsIncludeAllButK8s` function.

> If you want all options in `CreateOptions` to set to `true`  you may use `citfoptions.CreateOptionsIncludeAll` function.

CITF struct has four fields:- 
- Environment - To Setup or TearDown the platform such as minikube, GKE, AWS etc.
- K8S - K8S will have Kubernetes ClientSet & Config.
- Docker - Docker will be used for docker related operations.
- DebugEnabled - for verbose log.

> Currently CITF environment supports minikube only.

Developer can pass environment according to their requirements.

By default it will take Minikube as environment.

## Configuration

To configure the environment of CITF, there are three ways:-
 - [Environment Variable](#environment-variable)
 - [Config File](#config-file)
 - [Default Config](#default-config)

### Environment Variable

At the time of instantiation, developer can set CITF environment using environment variable `CITF_CONF_ENVIRONMENT`.

For example:- `export CITF_CONF_ENVIRONMENT = minikube`

### Config File
If environment variable is not set then developer can pass environment using config file. The file should be in `yaml` format. 

For example:- config.yaml

```
Environment: minikube
```

### Default Config

If environment variable and config file are not present, then CITF will take default environment which is minikube.

<details>
<summary><b>Platform Operations</b></summary>

`citf.Environment` will handle operations related to the platforms. 

In order to setup the k8s cluster, developer needs to call the `Setup()` method which will bring it up.

Developer can also check the status of the platform using `Status()` method.

Once integration test is completed, developer can delete the setup using `TearDown()` method.
</details>


<details>
<summary><b>Example</b></summary>

## example_test.go

```go
package example

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/CITF"
	citfoptions "github.com/openebs/CITF/citf_options"
)

var CitfInstance citf.CITF

func TestIntegrationExample(t *testing.T) {
	RegisterFailHandler(Fail)

	var err error
	// Initializing CITF without config file.
	// Also We should not include K8S as currently we don't have kubernetes environment setup
	CitfInstance, err = citf.NewCITF(citfoptions.CreateOptionsIncludeAllButK8s(""))
	Expect(err).NotTo(HaveOccurred())

	RunSpecs(t, "Integration Test Suite")
}

var _ = BeforeSuite(func() {

	// Setting up the default Platform i.e minikube
	err := CitfInstance.Environment.Setup()
	Expect(err).NotTo(HaveOccurred())

	// You have to update the K8s config when environment has been set up
	// this extra step will be unsolicited in upcoming changes.
	err = CitfInstance.Reload(citfoptions.CreateOptionsIncludeAll(""))
	Expect(err).NotTo(HaveOccurred())

	// Wait until platform is up
	time.Sleep(30 * time.Second)

	err = CitfInstance.K8S.YAMLApply("./nginx-rc.yaml")
	Expect(err).NotTo(HaveOccurred())

	// Wait until the pod is up and running
	time.Sleep(30 * time.Second)
})

var _ = AfterSuite(func() {

	// Tear Down the Platform
	err := CitfInstance.Environment.Teardown()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Integration Test", func() {
	When("We check the log", func() {
		It("has `started the controller` in the log", func() {
			pods, err := CitfInstance.K8S.GetPods("default", "nginx")
			Expect(err).NotTo(HaveOccurred())

			// Give pods some time to generate logs
			time.Sleep(2 * time.Second)

			// Assuming that only 1 nginx pod is running
			for _, v := range pods {
				log, err := CitfInstance.K8S.GetLog(v.GetName(), "default")
				Expect(err).NotTo(HaveOccurred())

				Expect(log).Should(ContainSubstring("started the controller"))
			}
		})
	})
})
```

Above example is using [Ginkgo](https://github.com/onsi/ginkgo) and [Gomega](https://github.com/onsi/gomega) framework for handling the tests.

`nginx-rc.yaml` which is used in above example is below.

## nginx-rc.yaml
```yaml
apiVersion: v1
kind: ReplicationController
metadata:
  name: nginx
spec:
  replicas: 1
  selector:
    app: nginx
  template:
    metadata:
      name: nginx
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        args: [/bin/sh, -c,
            'echo "started the controller"']
        ports:
        - containerPort: 80
```
> **Note:** Above yaml is compatible with kubernetes 1.9, you may need to modify it if your kubernetes version is different.

</details>
