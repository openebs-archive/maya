# list only our namespaced directories
PACKAGES = $(shell go list ./... | grep -v '/vendor/')

# Lint our code. Reference: https://golang.org/cmd/vet/
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods \
         -nilfunc -printf -rangeloops -shift -structtags -unsafeptr

# Tools required for different make targets or for development purposes
EXTERNAL_TOOLS=\
	github.com/Masterminds/glide \
	github.com/mitchellh/gox \
	golang.org/x/tools/cmd/cover \
	github.com/axw/gocov/gocov \
	gopkg.in/matm/v1/gocov-html \
	github.com/ugorji/go/codec/codecgen

# list only our .go files i.e. exlcudes any .go files from the vendor directory
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

all: test

dev: format
	@MAYA_DEV=1 sh -c "'$(PWD)/scripts/build.sh'"

bin:
	@sh -c "'$(PWD)/scripts/build.sh'"

clean: 
	rm -rf pkg
	rm -rf bin

release:
	@$(MAKE) bin

# Run the bootstrap target once before trying cov
cov:
	gocov test ./... | gocov-html > /tmp/coverage.html
	@cat /tmp/coverage.html

test:
	@echo "--> Running go fmt" ;
	@if [ -n "`go fmt ${PACKAGES}`" ]; then \
		echo "[ERR] go fmt updated formatting. Please commit formatted code first."; \
		exit 1; \
	fi
	@sh -c "'$(PWD)/scripts/test.sh'"
	@$(MAKE) vet

cover:
	go list ./... | grep -v vendor | xargs -n1 go test --cover

format:
	@echo "--> Running go fmt"
	@go fmt $(PACKAGES)

vet:
	@go tool vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
		go get golang.org/x/tools/cmd/vet; \
	fi
	@echo "--> Running go tool vet $(VETARGS) ${GOFILES_NOVENDOR}"
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
		go get $$tool; \
	done

# You might need to use sudo
install: bin/maya
	install -o root -g root -m 0755 ./bin/maya /usr/local/bin/maya

.PHONY: all bin cov integ test vet test-nodep
