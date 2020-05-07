# Specify the name for the binaries
CSPC_OPERATOR=cspc-operator
CSPC_OPERATOR_DEBUG=cspc-operator-debug

# Specify the name of the docker repo for amd64
CSPC_OPERATOR_REPO_NAME?=cspc-operator

.PHONY: cspc-operator-debug
cspc-operator-debug:
	@echo "----------------------------"
	@echo -n "--> ${CSPC_OPERATOR} "
	@echo "----------------------------"
	@PNAME=${CSPC_OPERATOR_DEBUG} CTLNAME=${CSPC_OPERATOR} BUILD_TAG="-tags=debug" sh -c "'$(PWD)/buildscripts/build.sh'"

.PHONY: cspc-operator-debug-image
cspc-operator-debug-image: cspc-operator-debug
	@echo "----------------------------"
	@echo -n "--> ${CSPC_OPERATOR} image "
	@echo "${IMAGE_ORG}/${CSPC_OPERATOR_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/${CSPC_OPERATOR_DEBUG}/${CSPC_OPERATOR} buildscripts/cspc-operator-debug/
	@cd buildscripts/${CSPC_OPERATOR_DEBUG} && sudo docker build -t ${IMAGE_ORG}/${CSPC_OPERATOR_REPO_NAME}:inject ${DBUILD_ARGS} .
	@rm buildscripts/${CSPC_OPERATOR_DEBUG}/${CSPC_OPERATOR}
