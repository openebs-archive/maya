# Specify the name of the docker repo for amd64
CVC_OPERATOR?=cvc-operator
# Specify the name of the docker repo for amd64
CVC_OPERATOR_REPO_NAME?=cvc-operator

# Specify the name of the docker repo for arm64
CVC_OPERATOR_REPO_NAME_ARM64?=cvc-operator-arm64

.PHONY: cvc-operator-image
cvc-operator-image:
	@echo "----------------------------"
	@echo -n "--> cvc-operator image "
	@echo "${HUB_USER}/${CVC_OPERATOR_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@PNAME=${CVC_OPERATOR} CTLNAME=${CVC_OPERATOR} sh -c "'$(PWD)/buildscripts/build.sh'"
	@cp bin/${CVC_OPERATOR}/${CVC_OPERATOR} buildscripts/cvc-operator/
	@cd buildscripts/${CVC_OPERATOR} && sudo docker build -t ${HUB_USER}/${CVC_OPERATOR_REPO_NAME}:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/${CVC_OPERATOR}/${CVC_OPERATOR}

.PHONY: cvc-operator-image.arm64
cvc-operator-image.arm64:
	@echo "----------------------------"
	@echo -n "--> arm64 based cstor-volume-mgmt image "
	@echo "${HUB_USER}/${CVC_OPERATOR_REPO_NAME_ARM64}:${IMAGE_TAG}"
	@echo "----------------------------"
	@PNAME=${CVC_OPERATOR} CTLNAME=${CVC_OPERATOR} sh -c "'$(PWD)/buildscripts/build.sh'"
	@cp bin/${CVC_OPERATOR}/${CVC_OPERATOR} buildscripts/cvc-operator/
	@cd buildscripts/${CVC_OPERATOR} && sudo docker build -t ${HUB_USER}/${CVC_OPERATOR_REPO_NAME_ARM64}:${IMAGE_TAG} -f Dockerfile.arm64 --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/${CVC_OPERATOR}/${CVC_OPERATOR}


