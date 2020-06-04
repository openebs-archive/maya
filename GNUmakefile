# Copyright © 2017 The OpenEBS Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

include buildscripts/common.mk

# list only the source code directories
PACKAGES = $(shell go list ./... | grep -v 'vendor\|pkg/client/generated\|tests')

# list only the integration tests code directories
PACKAGES_IT = $(shell go list ./... | grep -v 'vendor\|pkg/client/generated' | grep 'tests')

GO111MODULE ?= on
export GO111MODULE

# Lint our code. Reference: https://golang.org/cmd/vet/
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods \
         -nilfunc -printf -rangeloops -shift -structtags -unsafeptr

# API_PKG sets namespace where the API resources are defined
API_PKG := github.com/openebs/maya/pkg

# Default arguments for code gen script

# OUTPUT_PKG is the path of directory where you want to keep the generated code
OUTPUT_PKG=github.com/openebs/maya/pkg/client/generated

# APIS_PKG is the path where apis group and schema exists.
APIS_PKG=github.com/openebs/maya/pkg/apis

# GENS is an argument which generates different type of code.
# Possible values: all, deepcopy, client, informers, listers.
GENS=all
# GROUPS_WITH_VERSIONS is the group containing different versions of the resources.
GROUPS_WITH_VERSIONS=openebs.io:v1alpha1

# BOILERPLATE_TEXT_PATH is the boilerplate text(go comment) that is put at the top of every generated file.
# This boilerplate text is nothing but the license information.
BOILERPLATE_TEXT_PATH=buildscripts/custom-boilerplate.go.txt


# ALL_API_GROUPS has the list of all API resources from various groups
ALL_API_GROUPS=\
	openebs.io/runtask/v1beta1 \
	openebs.io/upgrade/v1alpha1 \
	openebs.io/snapshot/v1 \
	openebs.io/ndm/v1alpha1

# Tools required for different make targets or for development purposes
EXTERNAL_TOOLS=\
	github.com/golang/dep/cmd/dep \
	golang.org/x/tools/cmd/cover \
	github.com/axw/gocov/gocov \
	gopkg.in/matm/v1/gocov-html \
	github.com/ugorji/go/codec/codecgen \
	gopkg.in/alecthomas/gometalinter.v1 \
	github.com/golang/protobuf/protoc-gen-go

# list only our .go files i.e. exlcudes any .go files from the vendor directory
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

ifeq (${IMAGE_TAG}, )
  IMAGE_TAG = ci
  export IMAGE_TAG
endif

ifeq (${TRAVIS_TAG}, )
  BASE_TAG = ci
  export BASE_TAG
else
  BASE_TAG = $(TRAVIS_TAG:v%=%)
  export BASE_TAG
endif

# The images can be pushed to any docker/image registeries
# like docker hub, quay. The registries are specified in 
# the `build/push` script.
#
# The images of a project or company can then be grouped
# or hosted under a unique organization key like `openebs`
#
# Each component (container) will be pushed to a unique 
# repository under an organization. 
# Putting all this together, an unique uri for a given 
# image comprises of:
#   <registry url>/<image org>/<image repo>:<image-tag>
#
# IMAGE_ORG can be used to customize the organization 
# under which images should be pushed. 
# By default the organization name is `openebs`. 

ifeq (${IMAGE_ORG}, )
  IMAGE_ORG = openebs
  export IMAGE_ORG
endif

# Specify the date of build
DBUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Specify the docker arg for repository url
ifeq (${DBUILD_REPO_URL}, )
  DBUILD_REPO_URL="https://github.com/openebs/maya"
  export DBUILD_REPO_URL
endif

# Specify the docker arg for website url
ifeq (${DBUILD_SITE_URL}, )
  DBUILD_SITE_URL="https://openebs.io"
  export DBUILD_SITE_URL
endif

export DBUILD_ARGS=--build-arg DBUILD_DATE=${DBUILD_DATE} --build-arg DBUILD_REPO_URL=${DBUILD_REPO_URL} --build-arg DBUILD_SITE_URL=${DBUILD_SITE_URL} --build-arg ARCH=${ARCH}

# Specify the name of cstor-base image
CSTOR_BASE_IMAGE= ${IMAGE_ORG}/cstor-base:${BASE_TAG}
export CSTOR_BASE_IMAGE

ifeq (${CSTOR_BASE_IMAGE_ARM64}, )
  CSTOR_BASE_IMAGE_ARM64= ${IMAGE_ORG}/cstor-base-arm64:${BASE_TAG}
  export CSTOR_BASE_IMAGE_ARM64
endif

# Specify the name of base image for ARM64
ifeq (${BASE_DOCKER_IMAGE_ARM64}, )
  BASE_DOCKER_IMAGE_ARM64 = "arm64v8/ubuntu:18.04"
  export BASE_DOCKER_IMAGE_ARM64
endif

# Specify the name of base image for PPC64LE
ifeq (${BASE_DOCKER_IMAGE_PPC64LE}, )
  BASE_DOCKER_IMAGE_PPC64LE = "ubuntu:18.04"
  export BASE_DOCKER_IMAGE_PPC64LE
endif


include ./buildscripts/mayactl/Makefile.mk
include ./buildscripts/apiserver/Makefile.mk
include ./buildscripts/provisioner-localpv/Makefile.mk
include ./buildscripts/upgrade/Makefile.mk
include ./buildscripts/exporter/Makefile.mk
include ./buildscripts/cstor-pool-mgmt/Makefile.mk
include ./buildscripts/cstor-volume-mgmt/Makefile.mk
include ./buildscripts/cspi-mgmt/Makefile.mk
include ./buildscripts/cvc-operator/Makefile.mk
include ./buildscripts/admission-server/Makefile.mk
include ./buildscripts/cspc-operator/Makefile.mk
include ./buildscripts/cspc-operator-debug/Makefile.mk

.PHONY: all
all: deps compile-tests apiserver-image exporter-image pool-mgmt-image volume-mgmt-image \
	   admission-server-image cspc-operator-image cspc-operator-debug-image \
	   cvc-operator-image cspi-mgmt-image upgrade-image provisioner-localpv-image

.PHONY: all.arm64
all.arm64: apiserver-image.arm64 exporter-image.arm64 pool-mgmt-image.arm64 volume-mgmt-image.arm64 \
           admission-server-image.arm64 cspc-operator-image.arm64 upgrade-image.arm64 \
           cvc-operator-image.arm64 cspi-mgmt-image.arm64 provisioner-localpv-image.arm64

<<<<<<< HEAD
.PHONY: all.ppc64le
all.ppc64le: provisioner-localpv-image.ppc64le

.PHONY: initialize
initialize: bootstrap

.PHONY: deps
deps:
	@echo "--> Tidying up submodules"
	@go mod tidy
	@echo "--> Veryfying submodules"
	@go mod verify

.PHONY: verify-deps
verify-deps: deps
	@if !(git diff --quiet HEAD -- go.sum go.mod); then \
		echo "go module files are out of date, please commit the changes to go.mod and go.sum"; exit 1; \
	fi

.PHONY: clean
clean: cleanup-upgrade
	go clean -testcache
	rm -rf bin
	rm -rf ${GOPATH}/pkg/*

.PHONY: release
release:
	@$(MAKE) bin

# Run the bootstrap target once before trying cov
.PHONY: cov
cov:
	gocov test ./... | gocov-html > /tmp/coverage.html
	@cat /tmp/coverage.html

# Verifies compilation issues if any in integration test code
.PHONY: compile-tests
compile-tests:
	@echo "--> Running go vet on tests"
	@for test in  $(PACKAGES_IT) ; do \
		go vet $$test; \
	done

.PHONY: test
test: format
	@echo "--> Running go test" ;
	@go test $(PACKAGES)

.PHONY: testv
testv: format
	@echo "--> Running go test verbose" ;
	@go test -v $(PACKAGES)

.PHONY: cover
cover:
	go list ./... | grep -v vendor | xargs -n1 go test --cover

.PHONY: format
format:
	@echo "--> Running go fmt"
	@go fmt $(PACKAGES) $(PACKAGES_IT)

# Target to run gometalinter in Travis (deadcode, golint, errcheck, unconvert, goconst)
.PHONY: golint-travis
golint-travis:
	@gometalinter.v1 --install
	-gometalinter.v1 --config=metalinter.config ./...

# Run the bootstrap target once before trying gometalinter in Develop environment
.PHONY: golint
golint:
	@gometalinter.v1 --install
	@gometalinter.v1 --vendor --deadline=600s ./...

.PHONY: vet
vet:
	@go tool vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
		go get golang.org/x/tools/cmd/vet; \
	fi
	@echo "--> Running go tool vet ..."
	@go tool vet $(VETARGS) ${GOFILES_NOVENDOR} ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "[LINT] Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
	fi

	@git grep -n `echo "log"".Print"` | grep -v 'vendor/' ; if [ $$? -eq 0 ]; then \
		echo "[LINT] Found "log"".Printf" calls. These should use Maya's logger instead."; \
	fi

# Bootstrap the build by downloading additional tools
.PHONY: bootstrap
bootstrap:
	@for tool in  $(EXTERNAL_TOOLS) ; do \
		echo "+ Installing $$tool" ; \
		go get -u $$tool; \
	done

# code generation for custom resources
.PHONY: kubegen2
kubegen2: deepcopy2 clientset2 lister2 informer2

# code generation for custom resources
.PHONY: kubegen1
kubegen1:
	./buildscripts/code-gen.sh ${GENS} ${OUTPUT_PKG} ${APIS_PKG} ${GROUPS_WITH_VERSIONS} --go-header-file ${BOILERPLATE_TEXT_PATH}

# code generation for custom resources
.PHONY: kubegen
kubegen: kubegendelete kubegen1 kubegen2

# deletes generated code by codegen
.PHONY: kubegendelete
kubegendelete:
	@rm -rf pkg/client/generated/clientset
	@rm -rf pkg/client/generated/listers
	@rm -rf pkg/client/generated/informers
	@rm -rf pkg/client/generated/openebs.io

# code generation for custom resources and protobuf
.PHONY: generated_files
generated_files: kubegen protobuf

# builds vendored version of deepcopy-gen tool
.PHONY: deepcopy2
deepcopy2:
	@go install ./vendor/k8s.io/code-generator/cmd/deepcopy-gen
	@for apigrp in  $(ALL_API_GROUPS) ; do \
		echo "+ Generating deepcopy funcs for $$apigrp" ; \
		deepcopy-gen \
			--input-dirs $(API_PKG)/apis/$$apigrp \
			--output-file-base zz_generated.deepcopy \
			--go-header-file ./buildscripts/custom-boilerplate.go.txt; \
	done

# builds vendored version of client-gen tool
.PHONY: clientset2
clientset2:
	@go install ./vendor/k8s.io/code-generator/cmd/client-gen
	@for apigrp in  $(ALL_API_GROUPS) ; do \
		echo "+ Generating clientsets for $$apigrp" ; \
		client-gen \
			--fake-clientset=true \
			--input $$apigrp \
			--input-base $(API_PKG)/apis \
			--clientset-path $(API_PKG)/client/generated/$$apigrp/clientset \
			--go-header-file ./buildscripts/custom-boilerplate.go.txt; \
	done

# builds vendored version of lister-gen tool
.PHONY: lister2
lister2:
	@go install ./vendor/k8s.io/code-generator/cmd/lister-gen
	@for apigrp in  $(ALL_API_GROUPS) ; do \
		echo "+ Generating lister for $$apigrp" ; \
		lister-gen \
			--input-dirs $(API_PKG)/apis/$$apigrp \
			--output-package $(API_PKG)/client/generated/$$apigrp/lister \
			--go-header-file ./buildscripts/custom-boilerplate.go.txt; \
	done

# builds vendored version of informer-gen tool
.PHONY: informer2
informer2:
	@go install ./vendor/k8s.io/code-generator/cmd/informer-gen 
	@for apigrp in  $(ALL_API_GROUPS) ; do \
		echo "+ Generating informer for $$apigrp" ; \
		informer-gen \
			--input-dirs $(API_PKG)/apis/$$apigrp \
			--output-package $(API_PKG)/client/generated/$$apigrp/informer \
			--versioned-clientset-package $(API_PKG)/client/generated/$$apigrp/clientset/internalclientset \
			--listers-package $(API_PKG)/client/generated/$$apigrp/lister \
			--go-header-file ./buildscripts/custom-boilerplate.go.txt; \
	done

# You might need to use sudo
.PHONY: install
install: bin/maya/${MAYACTL}
	install -o root -g root -m 0755 ./bin/maya/${MAYACTL} /usr/local/bin/${MAYACTL}

# Push images
.PHONY: deploy-images
deploy-images:
	@./buildscripts/deploy.sh
