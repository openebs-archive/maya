# Makefile for setting up OpenEBS.
#
# Reference Guide - https://www.gnu.org/software/make/manual/make.html


#
# This is done to avoid conflict with a file of same name as the targets
# mentioned in this makefile.
#
.PHONY: help clean build install _build_check_go _clean_binaries

#
# Internal variables or constants.
# NOTE - These will be executed when any make target is invoked.
#
IS_GO_INSTALLED           := $(shell which go >> /dev/null 2>&1; echo $$?)

#
# The first target is the default.
# i.e. 'make' is same as 'make help'
#
help:
	@echo ""
	@echo "Usage:-"
	@echo -e "\tmake clean              -- will remove maya binaries from $(GOPATH)/bin"
	@echo -e "\tmake build              -- will build maya binaries"
	@echo -e "\tmake install            -- will build & install the maya binaries"
	@echo ""


#
# Will remove the openebs binaries at $GOPATH/bin
#
_clean_binaries:
	@echo ""
	@echo -e "INFO:\tremoving maya binaries from $(GOPATH)/bin ..."
	@rm -f $(GOPATH)/bin/maya
	@echo -e "INFO:\tmaya binaries removed successfully from $(GOPATH)/bin ..."
	@echo ""


#
# The clean target to be used by user.
#
clean: _clean_binaries

_build_check_go:
	@if [ $(IS_GO_INSTALLED) -eq 1 ]; \
		then echo "" \
		&& echo -e "ERROR:\tgo is not installed. Please install it before build." \
		&& echo -e "Refer:\thttps://github.com/openebs/maya#building-from-sources" \
		&& echo "" \
		&& exit 1; \
		fi;

#
# Will build the go based binaries
# The binaries will be placed at $GOPATH/bin/
#
build: _build_check_go
	@echo ""
	@echo -e "INFO:\tbuilding maya ..."
	@go get -t ./...
	@go get -u github.com/golang/lint/golint
	@echo -e "INFO:\tmaya built successfully ..."
	@echo ""

#
# Internally used target.
# Will place the maya binaries at /sbin/
#
_install_binary:
	@echo ""
	@echo -e "INFO:\tinstalling maya binaries ..."
	@cp $(GOPATH)/bin/maya /sbin/
	@echo -e "INFO:\tmaya binaries installed successfully ..."
	@echo ""


#
# The install target to be used by Admin.
#
install: _install_binary

