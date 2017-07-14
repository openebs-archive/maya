# list only our namespaced directories
PACKAGES = $(shell go list ./... | grep -v '/vendor/')

# Lint our code. Reference: https://golang.org/cmd/vet/
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods \
         -nilfunc -printf -rangeloops -shift -structtags -unsafeptr

# Tools required for different make targets or for development purposes
EXTERNAL_TOOLS=\
	github.com/kardianos/govendor \
	github.com/golang/lint/golint \
	github.com/mitchellh/gox \
	golang.org/x/tools/cmd/cover \
	github.com/axw/gocov/gocov \
	gopkg.in/matm/v1/gocov-html \
	github.com/ugorji/go/codec/codecgen

# list only our .go files i.e. exlcudes any .go files from the vendor directory
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Specify the name for the maya api server binary
CTLNAME=m-apiserver

all: test

dev: format
	@CTLNAME=${CTLNAME} M_APISERVER_DEV=1 sh -c "'$(PWD)/buildscripts/build.sh'"

bin:
	@CTLNAME=${CTLNAME} sh -c "'$(PWD)/buildscripts/build.sh'"

init: bootstrap deps

deps:
	rm -rf vendor/github.com/ && \
  rm -rf vendor/cloud.google.com/ && \
  rm -rf vendor/golang.org/ && \
  rm -rf vendor/gopkg.in/ && \
  rm -rf vendor/k8s.io/
	@echo "--> Sync with vendored repositories." ;
	@echo "--> Run this only when there is a change in vendor dependencies." ;
	@echo "--> Please wait, this may take a while..." ;
	@govendor sync

sync:
	@govendor sync

clean:
	rm -rf bin
	rm -rf ${GOPATH}/bin/${CTLNAME}

release:
	@$(MAKE) bin

# Run the bootstrap target once before trying cov
cov:
	gocov test ./... | gocov-html > /tmp/coverage.html
	@cat /tmp/coverage.html

test:
	@echo "--> Running go fmt" ;
	@if [ -n "`go fmt ${PACKAGES}`" ]; then \
		echo "[ERROR] go fmt has updated the formatting in some of the .go files."; \
		echo "--> If these files are open on any editor, then editor should reload & use the modified files."; \
		echo "--> Save these modified source files and proceed with this operation."; \
		echo "--> In some of the cases, these files will be auto saved. Hence, running this operation again will just work fine."; \
		exit 1; \
	fi
	@CTLNAME=${CTLNAME} sh -c "'$(PWD)/buildscripts/test.sh'"

cover:
	go list ./... | grep -v vendor | xargs -n1 go test --cover

format:
	@echo "--> Running go fmt"
	@go fmt $(PACKAGES)

lint:
	@echo "--> Running golint"
	@golint $(PACKAGES)
vet:
	@go tool vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
		go get golang.org/x/tools/cmd/vet; \
	fi
	@go tool vet $(VETARGS) ${GOFILES_NOVENDOR} ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "[LINT] Vet found suspicious constructs."; \
		echo "Fix them if necessary before submitting the code for review."; \
	fi

	@git grep -n `echo "log"".Print"` | grep -v 'vendor/' ; if [ $$? -eq 0 ]; then \
		echo "[LINT] Found "log"".Printf" calls. These should use Mayaserver's logger instead."; \
	fi

# Bootstrap the build by downloading additional tools
bootstrap:
	@for tool in  $(EXTERNAL_TOOLS) ; do \
		echo "Installing $$tool" ; \
		go get $$tool; \
	done 

# You might need to use sudo
install: bin/${CTLNAME}
	install -o root -g root -m 0755 ./bin/${CTLNAME} /usr/local/bin/${CTLNAME}

maya:
	go get github.com/openebs/maya
	ls ${GOPATH}/bin

image: maya
	@cp bin/m-apiserver buildscripts/docker/
	@cp ${GOPATH}/bin/maya buildscripts/docker/
	@cd buildscripts/docker && sudo docker build -t openebs/m-apiserver:ci .
	@sh buildscripts/push

.PHONY: all bin cov install test vet format cover bootstrap release clean deps init dev sync image
