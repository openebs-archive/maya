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

ifeq (${IMAGE_TAG}, )
  IMAGE_TAG = ci
  export IMAGE_TAG
endif

ifeq (${TRAVIS_TAG}, )
  BASE_TAG = develop-ci
  export BASE_TAG
else
  BASE_TAG = ${TRAVIS_TAG}
  export BASE_TAG
endif

CSTOR_BASE_IMAGE= openebs/cstor-base:${BASE_TAG}

# Specify the name for the binaries
MAYACTL=mayactl
APISERVER=maya-apiserver
WEBHOOK=admission-server
POOL_MGMT=cstor-pool-mgmt
VOLUME_MGMT=cstor-volume-mgmt
EXPORTER=maya-exporter
UPGRADE=upgrade

# Specify the date o build
BUILD_DATE = $(shell date +'%Y%m%d%H%M%S')

include ./buildscripts/provisioner-localpv/Makefile.mk

all: compile-tests mayactl apiserver-image exporter-image pool-mgmt-image volume-mgmt-image admission-server-image upgrade-image provisioner-localpv-image

mayactl:
	@echo "----------------------------"
	@echo "--> mayactl                    "
	@echo "----------------------------"
	@PNAME="maya" CTLNAME=${MAYACTL} sh -c "'$(PWD)/buildscripts/build.sh'"

initialize: bootstrap

deps:
	dep ensure

clean:
	go clean -testcache
	rm -rf bin
	rm -rf ${GOPATH}/bin/${MAYACTL}
	rm -rf ${GOPATH}/bin/${APISERVER}
	rm -rf ${GOPATH}/bin/${POOL_MGMT}
	rm -rf ${GOPATH}/bin/${VOLUME_MGMT}
	rm -rf ${GOPATH}/bin/${UPGRADE}
	rm -rf ${GOPATH}/pkg/*

release:
	@$(MAKE) bin

# Run the bootstrap target once before trying cov
cov:
	gocov test ./... | gocov-html > /tmp/coverage.html
	@cat /tmp/coverage.html

# Verifies compilation issues if any in integration test code
compile-tests:
	@echo "--> Running go vet on tests"
	@for test in  $(PACKAGES_IT) ; do \
		go vet $$test; \
	done

test: format
	@echo "--> Running go test" ;
	@go test $(PACKAGES)

testv: format
	@echo "--> Running go test verbose" ;
	@go test -v $(PACKAGES)

cover:
	go list ./... | grep -v vendor | xargs -n1 go test --cover

format:
	@echo "--> Running go fmt"
	@go fmt $(PACKAGES) $(PACKAGES_IT)

# Target to run gometalinter in Travis (deadcode, golint, errcheck, unconvert, goconst)
golint-travis:
	@gometalinter.v1 --install
	-gometalinter.v1 --config=metalinter.config ./...

# Run the bootstrap target once before trying gometalinter in Develop environment
golint:
	@gometalinter.v1 --install
	@gometalinter.v1 --vendor --deadline=600s ./...
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
bootstrap:
	@for tool in  $(EXTERNAL_TOOLS) ; do \
		echo "+ Installing $$tool" ; \
		go get -u $$tool; \
	done

# code generation for custom resources
kubegen2: deepcopy2 clientset2 lister2 informer2

# code generation for custom resources
kubegen1:
	./buildscripts/code-gen.sh ${GENS} ${OUTPUT_PKG} ${APIS_PKG} ${GROUPS_WITH_VERSIONS} --go-header-file ${BOILERPLATE_TEXT_PATH}

# code generation for custom resources
kubegen: kubegendelete kubegen1 kubegen2

# deletes generated code by codegen
kubegendelete:
	@rm -rf pkg/client/generated/clientset
	@rm -rf pkg/client/generated/listers
	@rm -rf pkg/client/generated/informers
	@rm -rf pkg/client/generated/openebs.io

# code generation for custom resources and protobuf
generated_files: kubegen protobuf

# builds vendored version of deepcopy-gen tool
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
install: bin/maya/${MAYACTL}
	install -o root -g root -m 0755 ./bin/maya/${MAYACTL} /usr/local/bin/${MAYACTL}

#Use this to build cstor-pool-mgmt
cstor-pool-mgmt:
	@echo "----------------------------"
	@echo "--> cstor-pool-mgmt           "
	@echo "----------------------------"
	@PNAME="cstor-pool-mgmt" CTLNAME=${POOL_MGMT} sh -c "'$(PWD)/buildscripts/build.sh'"

pool-mgmt-image: cstor-pool-mgmt
	@echo "----------------------------"
	@echo "--> cstor-pool-mgmt image "
	@echo "----------------------------"
	@cp bin/cstor-pool-mgmt/${POOL_MGMT} buildscripts/cstor-pool-mgmt/
	@cd buildscripts/cstor-pool-mgmt && sudo docker build -t openebs/cstor-pool-mgmt:${IMAGE_TAG} --build-arg BASE_IMAGE=${CSTOR_BASE_IMAGE} --build-arg BUILD_DATE=${BUILD_DATE} . --no-cache
	@rm buildscripts/cstor-pool-mgmt/${POOL_MGMT}

#Use this to build cstor-volume-mgmt
cstor-volume-mgmt:
	@echo "----------------------------"
	@echo "--> cstor-volume-mgmt           "
	@echo "----------------------------"
	@PNAME="cstor-volume-mgmt" CTLNAME=${VOLUME_MGMT} sh -c "'$(PWD)/buildscripts/build.sh'"

protobuf:
	@echo "----------------------------"
	@echo "--> protobuf           "
	@echo "----------------------------"
	@protoc -I $(PWD)/pkg/apis/openebs.io/v1alpha1/ \
    -I${GOPATH}/src \
    --go_out=plugins=grpc:$(PWD)/pkg/client/generated/cstor-volume-mgmt/v1alpha1 \
    $(PWD)/pkg/apis/openebs.io/v1alpha1/cstorvolume.proto

volume-mgmt-image: cstor-volume-mgmt
	@echo "----------------------------"
	@echo "--> cstor-volume-mgmt image         "
	@echo "----------------------------"
	@cp bin/cstor-volume-mgmt/${VOLUME_MGMT} buildscripts/cstor-volume-mgmt/
	@cd buildscripts/cstor-volume-mgmt && sudo docker build -t openebs/cstor-volume-mgmt:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/cstor-volume-mgmt/${VOLUME_MGMT}

# Use this to build only the maya-exporter.
exporter:
	@echo "----------------------------"
	@echo "--> maya-exporter              "
	@echo "----------------------------"
	@PNAME="exporter" CTLNAME=${EXPORTER} sh -c "'$(PWD)/buildscripts/build.sh'"

# m-exporter image. This is going to be decoupled soon.
exporter-image: exporter
	@echo "----------------------------"
	@echo "--> m-exporter image         "
	@echo "----------------------------"
	@cp bin/exporter/${EXPORTER} buildscripts/exporter/
	@cd buildscripts/exporter && sudo docker build -t openebs/m-exporter:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} --build-arg BASE_IMAGE=${CSTOR_BASE_IMAGE} .
	@rm buildscripts/exporter/${EXPORTER}

admission-server-image:
	@echo "----------------------------"
	@echo "--> admission-server image         "
	@echo "----------------------------"
	@PNAME=${WEBHOOK} CTLNAME=${WEBHOOK} sh -c "'$(PWD)/buildscripts/build.sh'"
	@cp bin/${WEBHOOK}/${WEBHOOK} buildscripts/admission-server/
	@cd buildscripts/${WEBHOOK} && sudo docker build -t openebs/admission-server:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/${WEBHOOK}/${WEBHOOK}

# Use this to build only the maya apiserver.
apiserver:
	@echo "----------------------------"
	@echo "--> maya-apiserver               "
	@echo "----------------------------"
	@PNAME="apiserver" CTLNAME=${APISERVER} sh -c "'$(PWD)/buildscripts/build.sh'"

# Currently both mayactl & apiserver binaries are pushed into
# m-apiserver image. This is going to be decoupled soon.
apiserver-image: mayactl apiserver
	@echo "----------------------------"
	@echo "--> apiserver image         "
	@echo "----------------------------"
	@cp bin/apiserver/${APISERVER} buildscripts/apiserver/
	@cp bin/maya/${MAYACTL} buildscripts/apiserver/
	@cd buildscripts/apiserver && sudo docker build -t openebs/m-apiserver:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/apiserver/${APISERVER}
	@rm buildscripts/apiserver/${MAYACTL}

rhel-apiserver-image: mayactl apiserver
	@echo "----------------------------"
	@echo "--> rhel based apiserver image"
	@echo "----------------------------"
	@cp bin/apiserver/${APISERVER} buildscripts/apiserver/
	@cp bin/maya/${MAYACTL} buildscripts/apiserver/
	@cd buildscripts/apiserver && sudo docker build -t openebs/m-apiserver:${IMAGE_TAG} -f Dockerfile.rhel --build-arg VERSION=${VERSION} .
	@rm buildscripts/apiserver/${APISERVER}
	@rm buildscripts/apiserver/${MAYACTL}

# Push images
deploy-images:
	@DIMAGE="openebs/m-apiserver" ./buildscripts/push
	@DIMAGE="openebs/m-exporter" ./buildscripts/push
	@DIMAGE="openebs/cstor-pool-mgmt" ./buildscripts/push
	@DIMAGE="openebs/cstor-volume-mgmt" ./buildscripts/push
	@DIMAGE="openebs/admission-server" ./buildscripts/push
	@DIMAGE="openebs/m-upgrade" ./buildscripts/push
	@DIMAGE="openebs/provisioner-localpv" ./buildscripts/push

# build upgrade binary
upgrade:
	@echo "----------------------------"
	@echo "--> ${UPGRADE}      "
	@echo "----------------------------"
	@PNAME=${UPGRADE} CTLNAME=${UPGRADE} CGO_ENABLED=0 sh -c "'$(PWD)/buildscripts/build.sh'"

# build upgrade image
upgrade-image: upgrade
	@echo "----------------------------"
	@echo "--> ${UPGRADE} image"
	@echo "----------------------------"
	@cp bin/${UPGRADE}/${UPGRADE} buildscripts/${UPGRADE}/
	@cd buildscripts/${UPGRADE} && sudo docker build -t openebs/m-upgrade:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/${UPGRADE}/${UPGRADE}

.PHONY: all bin cov integ test vet test-nodep apiserver image apiserver-image golint deploy kubegen kubegen2 generated_files deploy-images admission-server-image upgrade upgrade-image testv
