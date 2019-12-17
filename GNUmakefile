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
	openebs.io/snapshot/v1alpha1 \
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

# docker hub username
HUB_USER?=openebs

# Repository name
# format of docker image name is <hub-user>/<repo-name>[:<tag>].
# so final name will be ${HUB_USER}/${*_REPO_NAME}:${IMAGE_TAG}
CSTOR_POOL_MGMT_REPO_NAME?=cstor-pool-mgmt
CSPI_MGMT_REPO_NAME?=cspi-mgmt
CSTOR_VOLUME_MGMT_REPO_NAME?=cstor-volume-mgmt
M_EXPORTER_REPO_NAME?=m-exporter
ADMISSION_SERVER_REPO_NAME?=admission-server
M_UPGRADE_REPO_NAME?=m-upgrade
CSPC_OPERATOR_REPO_NAME?=cspc-operator

ifeq (${IMAGE_TAG}, )
  IMAGE_TAG = ci
  export IMAGE_TAG
endif

ifeq (${TRAVIS_TAG}, )
  BASE_TAG = ci
  export BASE_TAG
else
  BASE_TAG = ${TRAVIS_TAG}
  export BASE_TAG
endif

CSTOR_BASE_IMAGE= openebs/cstor-base:${BASE_TAG}

ifeq (${BASE_DOCKER_IMAGEARM64}, )
  BASE_DOCKER_IMAGEARM64 = "arm64v8/ubuntu:18.04"
  export BASE_DOCKER_IMAGEARM64
endif

# Specify the name for the binaries
WEBHOOK=admission-server
POOL_MGMT=cstor-pool-mgmt
CSPI_MGMT=cspi-mgmt
VOLUME_MGMT=cstor-volume-mgmt
EXPORTER=maya-exporter
CSPC_OPERATOR=cspc-operator
CSPC_OPERATOR_DEBUG=cspc-operator-debug
CSP_OPERATOR_DEBUG=cstor-pool-mgmt-debug


# Specify the date o build
BUILD_DATE = $(shell date +'%Y%m%d%H%M%S')

include ./buildscripts/mayactl/Makefile.mk
include ./buildscripts/apiserver/Makefile.mk
include ./buildscripts/provisioner-localpv/Makefile.mk
include ./buildscripts/upgrade/Makefile.mk
include ./buildscripts/upgrade-082090/Makefile.mk

.PHONY: all
all: compile-tests apiserver-image exporter-image pool-mgmt-image volume-mgmt-image admission-server-image cspc-operator-image cspc-operator-debug-image cspi-mgmt-image upgrade-image provisioner-localpv-image

.PHONY: all.arm64
all.arm64: apiserver-image.arm64 provisioner-localpv-image.arm64

.PHONY: initialize
initialize: bootstrap

.PHONY: deps
deps:
	dep ensure

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

#Use this to build cstor-pool-mgmt
.PHONY: cstor-pool-mgmt
cstor-pool-mgmt:
	@echo "----------------------------"
	@echo "--> cstor-pool-mgmt           "
	@echo "----------------------------"
	@PNAME="cstor-pool-mgmt" CTLNAME=${POOL_MGMT} sh -c "'$(PWD)/buildscripts/build.sh'"

.PHONY: pool-mgmt-image
pool-mgmt-image: cstor-pool-mgmt
	@echo "----------------------------"
	@echo -n "--> cstor-pool-mgmt image "
	@echo "${HUB_USER}/${CSTOR_POOL_MGMT_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/cstor-pool-mgmt/${POOL_MGMT} buildscripts/cstor-pool-mgmt/
	@cd buildscripts/cstor-pool-mgmt && sudo docker build -t ${HUB_USER}/${CSTOR_POOL_MGMT_REPO_NAME}:${IMAGE_TAG} --build-arg BASE_IMAGE=${CSTOR_BASE_IMAGE} --build-arg BUILD_DATE=${BUILD_DATE} . --no-cache
	@rm buildscripts/cstor-pool-mgmt/${POOL_MGMT}

#Use this to build debug image of cstor-pool-mgmt
.PHONY: pool-mgmt-debug-image
pool-mgmt-debug-image:
	@echo "----------------------------"
	@echo -n "--> cstor-pool-mgmt debug image "
	@echo "${HUB_USER}/${CSTOR_POOL_MGMT_REPO_NAME}:inject"
	@echo "----------------------------"
	@PNAME=${CSP_OPERATOR_DEBUG} CTLNAME=${POOL_MGMT} BUILD_TAG="-tags=debug" sh -c "'$(PWD)/buildscripts/build.sh'"
	@cp bin/${CSP_OPERATOR_DEBUG}/${POOL_MGMT} buildscripts/${CSP_OPERATOR_DEBUG}/
	@cd buildscripts/${CSP_OPERATOR_DEBUG} && sudo docker build -t ${HUB_USER}/${CSTOR_POOL_MGMT_REPO_NAME}:inject --build-arg BASE_IMAGE=${CSTOR_BASE_IMAGE} --build-arg BUILD_DATE=${BUILD_DATE} . --no-cache
	@rm buildscripts/${CSP_OPERATOR_DEBUG}/${POOL_MGMT}

#Use this to build cspi-mgmt
.PHONY: cspi-mgmt
cspi-mgmt:
	@echo "----------------------------"
	@echo "--> cspi-mgmt           "
	@echo "----------------------------"
	@PNAME="cspi-mgmt" CTLNAME=${CSPI_MGMT} sh -c "'$(PWD)/buildscripts/build.sh'"

.PHONY: cspi-mgmt-image
cspi-mgmt-image: cspi-mgmt
	@echo "----------------------------"
	@echo -n "--> cspi-mgmt image "
	@echo "${HUB_USER}/${CSPI_MGMT_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/cspi-mgmt/${CSPI_MGMT} buildscripts/cspi-mgmt/
	@cd buildscripts/cspi-mgmt && sudo docker build -t ${HUB_USER}/${CSPI_MGMT_REPO_NAME}:${IMAGE_TAG} --build-arg BASE_IMAGE=${CSTOR_BASE_IMAGE} --build-arg BUILD_DATE=${BUILD_DATE} . --no-cache
	@rm buildscripts/cspi-mgmt/${CSPI_MGMT}

#Use this to build cstor-volume-mgmt
.PHONY: cstor-volume-mgmt
cstor-volume-mgmt:
	@echo "----------------------------"
	@echo "--> cstor-volume-mgmt           "
	@echo "----------------------------"
	@PNAME="cstor-volume-mgmt" CTLNAME=${VOLUME_MGMT} sh -c "'$(PWD)/buildscripts/build.sh'"

.PHONY: protobuf
protobuf:
	@echo "----------------------------"
	@echo "--> protobuf           "
	@echo "----------------------------"
	@protoc -I $(PWD)/pkg/apis/openebs.io/v1alpha1/ \
    -I${GOPATH}/src \
    --go_out=plugins=grpc:$(PWD)/pkg/client/generated/cstor-volume-mgmt/v1alpha1 \
    $(PWD)/pkg/apis/openebs.io/v1alpha1/cstorvolume.proto

.PHONY: volume-mgmt-image
volume-mgmt-image: cstor-volume-mgmt
	@echo "----------------------------"
	@echo -n "--> cstor-volume-mgmt image "
	@echo "${HUB_USER}/${CSTOR_VOLUME_MGMT_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/cstor-volume-mgmt/${VOLUME_MGMT} buildscripts/cstor-volume-mgmt/
	@cd buildscripts/cstor-volume-mgmt && sudo docker build -t ${HUB_USER}/${CSTOR_VOLUME_MGMT_REPO_NAME}:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/cstor-volume-mgmt/${VOLUME_MGMT}

# Use this to build only the maya-exporter.
.PHONY: exporter
exporter:
	@echo "----------------------------"
	@echo "--> maya-exporter              "
	@echo "----------------------------"
	@PNAME="exporter" CTLNAME=${EXPORTER} sh -c "'$(PWD)/buildscripts/build.sh'"

# m-exporter image. This is going to be decoupled soon.
.PHONY: exporter-image
exporter-image: exporter
	@echo "----------------------------"
	@echo -n "--> m-exporter image "
	@echo "${HUB_USER}/${M_EXPORTER_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/exporter/${EXPORTER} buildscripts/exporter/
	@cd buildscripts/exporter && sudo docker build -t ${HUB_USER}/${M_EXPORTER_REPO_NAME}:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} --build-arg BASE_IMAGE=${CSTOR_BASE_IMAGE} .
	@rm buildscripts/exporter/${EXPORTER}

.PHONY: admission-server-image
admission-server-image:
	@echo "----------------------------"
	@echo -n "--> admission-server image "
	@echo "${HUB_USER}/${ADMISSION_SERVER_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@PNAME=${WEBHOOK} CTLNAME=${WEBHOOK} sh -c "'$(PWD)/buildscripts/build.sh'"
	@cp bin/${WEBHOOK}/${WEBHOOK} buildscripts/admission-server/
	@cd buildscripts/${WEBHOOK} && sudo docker build -t ${HUB_USER}/${ADMISSION_SERVER_REPO_NAME}:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/${WEBHOOK}/${WEBHOOK}

.PHONY: cspc-operator-image
cspc-operator-image:
	@echo "----------------------------"
	@echo -n "--> cspc-operator image "
	@echo "${HUB_USER}/${CSPC_OPERATOR_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@PNAME=${CSPC_OPERATOR} CTLNAME=${CSPC_OPERATOR} sh -c "'$(PWD)/buildscripts/build.sh'"
	@cp bin/${CSPC_OPERATOR}/${CSPC_OPERATOR} buildscripts/cspc-operator/
	@cd buildscripts/${CSPC_OPERATOR} && sudo docker build -t ${HUB_USER}/${CSPC_OPERATOR_REPO_NAME}:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/${CSPC_OPERATOR}/${CSPC_OPERATOR}

.PHONY: cspc-operator-debug-image
cspc-operator-debug-image:
	@echo "----------------------------"
	@echo -n "--> cspc-operator image "
	@echo "${HUB_USER}/${CSPC_OPERATOR_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@PNAME=${CSPC_OPERATOR_DEBUG} CTLNAME=${CSPC_OPERATOR} BUILD_TAG="-tags=debug" sh -c "'$(PWD)/buildscripts/build.sh'"
	@cp bin/${CSPC_OPERATOR_DEBUG}/${CSPC_OPERATOR} buildscripts/cspc-operator-debug/
	@cd buildscripts/${CSPC_OPERATOR_DEBUG} && sudo docker build -t ${HUB_USER}/${CSPC_OPERATOR_REPO_NAME}:inject --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/${CSPC_OPERATOR_DEBUG}/${CSPC_OPERATOR}

# Push images
.PHONY: deploy-images
deploy-images:
	@DIMAGE="openebs/m-apiserver" ./buildscripts/push
	@DIMAGE="openebs/m-exporter" ./buildscripts/push
	@DIMAGE="openebs/cstor-pool-mgmt" ./buildscripts/push
	@DIMAGE="openebs/cspi-mgmt" ./buildscripts/push
	@DIMAGE="openebs/cstor-volume-mgmt" ./buildscripts/push
	@DIMAGE="openebs/admission-server" ./buildscripts/push
	@DIMAGE="openebs/cspc-operator" ./buildscripts/push
	@DIMAGE="${HUB_USER}/${M_UPGRADE_REPO_NAME}" ./buildscripts/push
	@DIMAGE="openebs/provisioner-localpv" ./buildscripts/push
