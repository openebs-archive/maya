
# Specify the name for the binaries
CSPC_OPERATOR=cspc-operator

# Specify the name of the docker repo for amd64
CSPC_OPERATOR_REPO_NAME?=cspc-operator

# Specify the name of the docker repo for arm64
CSPC_OPERATOR_REPO_NAME_ARM64?=cspc-operator-arm64

.PHONY: cspc-operator
cspc-operator:
	@echo "----------------------------"
	@echo -n "--> ${CSPC_OPERATOR} "
	@echo "----------------------------"
	@PNAME=${CSPC_OPERATOR} CTLNAME=${CSPC_OPERATOR} sh -c "'$(PWD)/buildscripts/build.sh'"

.PHONY: cspc-operator-image
cspc-operator-image: cspc-operator
	@echo "----------------------------"
	@echo -n "--> ${CSPC_OPERATOR} image "
	@echo "${IMAGE_ORG}/${CSPC_OPERATOR_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/${CSPC_OPERATOR}/${CSPC_OPERATOR} buildscripts/cspc-operator/
	@cd buildscripts/${CSPC_OPERATOR} && sudo docker build -t ${IMAGE_ORG}/${CSPC_OPERATOR_REPO_NAME}:${IMAGE_TAG} ${DBUILD_ARGS} .
	@rm buildscripts/${CSPC_OPERATOR}/${CSPC_OPERATOR}

.PHONY: cspc-operator-image.arm64
cspc-operator-image.arm64: cspc-operator
	@echo "----------------------------"
	@echo -n "--> ${CSPC_OPERATOR} image "
	@echo "${IMAGE_ORG}/${CSPC_OPERATOR_REPO_NAME_ARM64}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/${CSPC_OPERATOR}/${CSPC_OPERATOR} buildscripts/cspc-operator/
	@cd buildscripts/${CSPC_OPERATOR} && sudo docker build -t ${IMAGE_ORG}/${CSPC_OPERATOR_REPO_NAME_ARM64}:${IMAGE_TAG} -f Dockerfile.arm64 ${DBUILD_ARGS} .
	@rm buildscripts/${CSPC_OPERATOR}/${CSPC_OPERATOR}
