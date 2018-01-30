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

# Specify the name for the binaries
MAYACTL=mayactl
APISERVER=maya-apiserver
AGENT=maya-agent
EXPORTER=maya-volume-exporter

# Specify the date o build
BUILD_DATE = $(shell date +'%Y%m%d%H%M%S')

all: mayactl apiserver-image exporter-image maya-agent

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
	rm -rf ${GOPATH}/bin/${MAYACTL}
	rm -rf ${GOPATH}/bin/${APISERVER}
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
	@cd buildscripts/mayactl && sudo docker build -t openebs/maya:ci --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/mayactl/${MAYACTL}
	@sh buildscripts/mayactl/push

# You might need to use sudo
install: bin/maya/${MAYACTL}
	install -o root -g root -m 0755 ./bin/maya/${MAYACTL} /usr/local/bin/${MAYACTL}

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
	@cd buildscripts/agent && sudo docker build -t openebs/m-agent:ci --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/agent/${AGENT}
	@sh buildscripts/agent/push

# Use this to build only the maya-volume-exporter.
exporter:
	@echo "----------------------------"
	@echo "--> maya-volume-exporter              "
	@echo "----------------------------"
	@CTLNAME=${EXPORTER} sh -c "'$(PWD)/buildscripts/exporter/build.sh'"

# m-exporter image. This is going to be decoupled soon.
exporter-image: exporter
	@echo "----------------------------"
	@echo "--> m-exporter image         "
	@echo "----------------------------"
	@cp bin/exporter/${EXPORTER} buildscripts/exporter/
	@cd buildscripts/exporter && sudo docker build -t openebs/m-exporter:ci --build-arg BUILD_DATE=${BUILD_DATE} .
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
	@cd buildscripts/apiserver && sudo docker build -t openebs/m-apiserver:ci --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/apiserver/${APISERVER}
	@rm buildscripts/apiserver/${MAYACTL}
	@sh buildscripts/apiserver/push

.PHONY: all bin cov integ test vet maya-agent test-nodep apiserver image apiserver-image maya-image golint
