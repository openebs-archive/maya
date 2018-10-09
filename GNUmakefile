# list only our namespaced directories
PACKAGES = $(shell go list ./... | grep -v 'vendor\|pkg/apis\|pkg/client/generated')

# Lint our code. Reference: https://golang.org/cmd/vet/
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods \
         -nilfunc -printf -rangeloops -shift -structtags -unsafeptr

# API_PKG sets namespace where the API resources are defined
API_PKG := github.com/openebs/maya/pkg

# API_GROUPS sets api version of the resources exposed by maya
ifeq (${API_GROUPS}, )
  API_GROUPS = openebs.io/v1alpha1
  export API_GROUPS
endif

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
  BASE_TAG = ${TRAVIS_TAG}
  export BASE_TAG
endif

CSTOR_BASE_IMAGE= openebs/cstor-base:${BASE_TAG}

# Specify the name for the binaries
MAYACTL=mayactl
APISERVER=maya-apiserver
POOL_MGMT=cstor-pool-mgmt
VOLUME_MGMT=cstor-volume-mgmt
VOLUME_GRPC=cstor-volume-grpc
EXPORTER=maya-exporter

# Specify the date o build
BUILD_DATE = $(shell date +'%Y%m%d%H%M%S')

all: mayactl apiserver-image exporter-image pool-mgmt-image volume-mgmt-image

dev: format
	@MAYACTL=${MAYACTL} MAYA_DEV=1 sh -c "'$(PWD)/buildscripts/build.sh'" maya

mayactl:
	@echo "----------------------------"
	@echo "--> mayactl                    "
	@echo "----------------------------"
	@MAYACTL=${MAYACTL} sh -c "'$(PWD)/buildscripts/build.sh'" maya

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
	rm -rf ${GOPATH}/bin/${VOLUME_GRPC}
	rm -rf ${GOPATH}/pkg/*

release:
	@$(MAKE) bin

# Run the bootstrap target once before trying cov
cov:
	gocov test ./... | gocov-html > /tmp/coverage.html
	@cat /tmp/coverage.html

test: format
	@echo "--> Running go test" ;
	@go test $(PACKAGES)

cover:
	go list ./... | grep -v vendor | xargs -n1 go test --cover

format:
	@echo "--> Running go fmt"
	@go fmt $(PACKAGES)

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
		echo "Installing $$tool" ; \
		go get -u $$tool; \
	done

# code generation for custom resources
generated_files: deepcopy clientset lister informer cstor-volume-grpc

# builds vendored version of deepcopy-gen tool
deepcopy:
	@go install ./vendor/k8s.io/code-generator/cmd/deepcopy-gen
	@echo "+ Generating deepcopy funcs for $(API_GROUPS)"
	@deepcopy-gen \
		--input-dirs $(API_PKG)/apis/$(API_GROUPS) \
		--output-file-base zz_generated.deepcopy \
		--go-header-file ./buildscripts/custom-boilerplate.go.txt

# also builds vendored version of client-gen tool
clientset:
	@go install ./vendor/k8s.io/code-generator/cmd/client-gen
	@echo "+ Generating clientsets for $(API_GROUPS)"
	@client-gen \
		--fake-clientset=true \
		--input $(API_GROUPS) \
		--input-base $(API_PKG)/apis \
		--clientset-path $(API_PKG)/client/generated/clientset \
		--go-header-file ./buildscripts/custom-boilerplate.go.txt

# also builds vendored version of lister-gen tool
lister:
	@go install ./vendor/k8s.io/code-generator/cmd/lister-gen
	@echo "+ Generating lister for $(API_GROUPS)"
	@lister-gen \
		--input-dirs $(API_PKG)/apis/$(API_GROUPS) \
		--output-package $(API_PKG)/client/generated/lister \
		--go-header-file ./buildscripts/custom-boilerplate.go.txt

# also builds vendored version of informer-gen tool
informer:
	@go install ./vendor/k8s.io/code-generator/cmd/informer-gen
	@echo "+ Generating informer for $(API_GROUPS)"
	@informer-gen \
		--input-dirs $(API_PKG)/apis/$(API_GROUPS) \
		--output-package $(API_PKG)/client/generated/informer \
		--versioned-clientset-package $(API_PKG)/client/generated/clientset/internalclientset \
		--listers-package $(API_PKG)/client/generated/lister \
		--go-header-file ./buildscripts/custom-boilerplate.go.txt

# You might need to use sudo
install: bin/maya/${MAYACTL}
	install -o root -g root -m 0755 ./bin/maya/${MAYACTL} /usr/local/bin/${MAYACTL}

#Use this to build cstor-pool-mgmt
cstor-pool-mgmt:
	@echo "----------------------------"
	@echo "--> cstor-pool-mgmt           "
	@echo "----------------------------"
	@CTLNAME=${POOL_MGMT} sh -c "'$(PWD)/buildscripts/cstor-pool-mgmt/build.sh'"

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
	@CTLNAME=${VOLUME_MGMT} sh -c "'$(PWD)/buildscripts/cstor-volume-mgmt/build.sh'"

#Use this to build cstor-volume-grpc
cstor-volume-grpc:
	@echo "----------------------------"
	@echo "--> cstor-volume-grpc           "
	@echo "----------------------------"
	@protoc -I $(PWD)/pkg/apis/openebs.io/v1alpha1/ \
    -I${GOPATH}/src \
    --go_out=plugins=grpc:$(PWD)/pkg/client/generated/cstor-volume-grpc/v1alpha1 \
    $(PWD)/pkg/apis/openebs.io/v1alpha1/cstorvolume.proto

volume-mgmt-image: cstor-volume-mgmt
	@echo "----------------------------"
	@echo "--> cstor-volume-mgmt image         "
	@echo "----------------------------"
	@cp bin/cstor-volume-mgmt/${VOLUME_MGMT} buildscripts/cstor-volume-mgmt/
	@cp bin/cstor-volume-mgmt/${VOLUME_GRPC} buildscripts/cstor-volume-mgmt/
	@cd buildscripts/cstor-volume-mgmt && sudo docker build -t openebs/cstor-volume-mgmt:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/cstor-volume-mgmt/${VOLUME_MGMT}
	@rm buildscripts/cstor-volume-mgmt/${VOLUME_GRPC}

# Use this to build only the maya-exporter.
exporter:
	@echo "----------------------------"
	@echo "--> maya-exporter              "
	@echo "----------------------------"
	@CTLNAME=${EXPORTER} sh -c "'$(PWD)/buildscripts/build.sh'" exporter

# m-exporter image. This is going to be decoupled soon.
exporter-image: exporter
	@echo "----------------------------"
	@echo "--> m-exporter image         "
	@echo "----------------------------"
	@cp bin/exporter/${EXPORTER} buildscripts/exporter/
	@cd buildscripts/exporter && sudo docker build -t openebs/m-exporter:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/exporter/${EXPORTER}

# Use this to build only the maya apiserver.
apiserver:
	@echo "----------------------------"
	@echo "--> maya-apiserver               "
	@echo "----------------------------"
	@CTLNAME=${APISERVER} sh -c "'$(PWD)/buildscripts/build.sh'" apiserver

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

# Push images
deploy-images:
	@DIMAGE="openebs/m-apiserver" ./buildscripts/push
	@DIMAGE="openebs/m-exporter" ./buildscripts/push
	@DIMAGE="openebs/cstor-pool-mgmt" ./buildscripts/push
	@DIMAGE="openebs/cstor-volume-mgmt" ./buildscripts/push

.PHONY: all bin cov integ test vet test-nodep apiserver image apiserver-image golint deploy
