# Makefile for CITF

build: vet fmt golint

# Tools required for different make targets or for development purposes
EXTERNAL_TOOLS = \
	github.com/fzipp/gocyclo \
	golang.org/x/lint/golint \
	github.com/onsi/ginkgo/ginkgo \
	github.com/onsi/gomega/...

vet:
	go list ./... | grep -v "./vendor/*" | xargs go vet

fmt:
	find . -type f -name "*.go" | grep -v "./vendor/*" | xargs gofmt -s -w -l

# Run the bootstrap target once before trying golint in Development environment
golint:
	go list ./... | grep -v "./vendor/*" | xargs golint

# Run the bootstrap target once before trying gocyclo in Development environment
gocyclo:
	gocyclo . | grep -v vendor

# Target for running go test
test: vet fmt
	@echo "--> Running go test";
	$(PWD)/test.sh

integration-test:
	go test -v github.com/openebs/CITF/example

# Bootstrap the build by downloading additional tools
bootstrap:
	@for tool in $(EXTERNAL_TOOLS) ; do \
		echo "Installing $$tool" ; \
		go get -u $$tool; \
	done

.PHONY: build
