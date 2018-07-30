# list only our namespaced directories
PACKAGES = $(shell go list ./... | grep -v '/vendor/')

# Lint our code. Reference: https://golang.org/cmd/vet/
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods \
         -nilfunc -printf -rangeloops -shift -structtags -unsafeptr

# Tools required for different make targets or for development purposes
EXTERNAL_TOOLS=\
	github.com/golang/dep/cmd/dep \
	golang.org/x/tools/cmd/cover \
	github.com/axw/gocov/gocov \
	gopkg.in/matm/v1/gocov-html \
	github.com/ugorji/go/codec/codecgen \
	gopkg.in/alecthomas/gometalinter.v1

# list only our .go files i.e. exlcudes any .go files from the vendor directory
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

ifeq (${IMAGE_TAG}, )
  IMAGE_TAG = local
  export IMAGE_TAG
endif

# Specify the name for the binaries
MAYACTL=mayactl
APISERVER=maya-apiserver
POOL_MGMT=cstor-pool-mgmt
VOLUME_MGMT=cstor-volume-mgmt
VOLUME_GRPC=cstor-volume-grpc
AGENT=maya-agent
EXPORTER=maya-exporter

# Specify the date o build
BUILD_DATE = $(shell date +'%Y%m%d%H%M%S')

all: mayactl apiserver-image exporter-image maya-agent pool-mgmt-image volume-mgmt-image

dev: format
	@MAYACTL=${MAYACTL} MAYA_DEV=1 sh -c "'$(PWD)/buildscripts/mayactl/build.sh'"

mayactl:
	@echo "----------------------------"
	@echo "--> mayactl                    "
	@echo "----------------------------"
	@MAYACTL=${MAYACTL} sh -c "'$(PWD)/buildscripts/mayactl/build.sh'"

initialize: bootstrap

deps:
	dep ensure

clean:
	rm -rf bin
	rm -rf cmd/cstor-volume-grpc/api/*.pb.go
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

maya-image:
	@cp bin/maya/${MAYACTL} buildscripts/mayactl/
	@cd buildscripts/mayactl && sudo docker build -t openebs/maya:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/mayactl/${MAYACTL}
	@sh buildscripts/mayactl/push

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
	@echo "--> cstor-pool-mgmt image         "
	@echo "----------------------------"
	@cp bin/cstor-pool-mgmt/${POOL_MGMT} buildscripts/cstor-pool-mgmt/
	@cd buildscripts/cstor-pool-mgmt && sudo docker build -t openebs/cstor-pool-mgmt:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} . --no-cache
	@rm buildscripts/cstor-pool-mgmt/${POOL_MGMT}
	@sh buildscripts/cstor-pool-mgmt/push

#Use this to build cstor-volume-mgmt
cstor-volume-mgmt: cstor-volume-grpc
	@echo "----------------------------"
	@echo "--> cstor-volume-mgmt           "            
	@echo "----------------------------"
	@CTLNAME=${VOLUME_MGMT} sh -c "'$(PWD)/buildscripts/cstor-volume-mgmt/build.sh'"

#Use this to build cstor-volume-grpc
cstor-volume-grpc:
	@echo "----------------------------"
	@echo "--> cstor-volume-grpc           "            
	@echo "----------------------------"
	@protoc -I $(PWD)/cmd/cstor-volume-grpc/api/ \
    -I${GOPATH}/src \
    --go_out=plugins=grpc:$(PWD)/cmd/cstor-volume-grpc/api \
    $(PWD)/cmd/cstor-volume-grpc/api/api.proto

volume-mgmt-image: cstor-volume-mgmt
	@echo "----------------------------"
	@echo "--> cstor-volume-mgmt image         "
	@echo "----------------------------"
	@cp bin/cstor-volume-mgmt/${VOLUME_MGMT} buildscripts/cstor-volume-mgmt/
	@cp bin/cstor-volume-mgmt/${VOLUME_GRPC} buildscripts/cstor-volume-mgmt/
	@cd buildscripts/cstor-volume-mgmt && sudo docker build -t openebs/cstor-volume-mgmt:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/cstor-volume-mgmt/${VOLUME_MGMT}
	@rm buildscripts/cstor-volume-mgmt/${VOLUME_GRPC}
	@sh buildscripts/cstor-volume-mgmt/push

# Use this to build only the maya-agent.
maya-agent:
	@echo "----------------------------"
	@echo "--> maya-agent              "
	@echo "----------------------------"
	@CTLNAME=${AGENT} sh -c "'$(PWD)/buildscripts/agent/build.sh'"

# m-agent image. This is going to be decoupled soon.
agent-image: maya-agent
	@echo "----------------------------"
	@echo "--> m-agent image         "
	@echo "----------------------------"
	@cp bin/agent/${AGENT} buildscripts/agent/
	@cd buildscripts/agent && sudo docker build -t openebs/m-agent:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/agent/${AGENT}
	@sh buildscripts/agent/push

# Use this to build only the maya-exporter.
exporter:
	@echo "----------------------------"
	@echo "--> maya-exporter              "
	@echo "----------------------------"
	@CTLNAME=${EXPORTER} sh -c "'$(PWD)/buildscripts/exporter/build.sh'"

# m-exporter image. This is going to be decoupled soon.
exporter-image: exporter
	@echo "----------------------------"
	@echo "--> m-exporter image         "
	@echo "----------------------------"
	@cp bin/exporter/${EXPORTER} buildscripts/exporter/
	@cd buildscripts/exporter && sudo docker build -t openebs/m-exporter:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/exporter/${EXPORTER}
	@sh buildscripts/exporter/push

# Use this to build only the maya apiserver.
apiserver:
	@echo "----------------------------"
	@echo "--> maya-apiserver               "
	@echo "----------------------------"
	@CTLNAME=${APISERVER} sh -c "'$(PWD)/buildscripts/apiserver/build.sh'"

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
	@sh buildscripts/apiserver/push

.PHONY: all bin cov integ test vet maya-agent test-nodep apiserver image apiserver-image maya-image golint
