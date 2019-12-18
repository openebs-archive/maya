
# Specify the name for the binaries
POOL_MGMT=cstor-pool-mgmt
CSP_OPERATOR_DEBUG=csotr-pool-mgmt-debug

#Specify the name of the docker repo for amd64
CSTOR_POOL_MGMT_REPO_NAME?=cstor-pool-mgmt

#Specify the name of the docker repo for arm64
CSTOR_POOL_MGMT_REPO_NAME_ARM64?=cstor-pool-mgmt-arm64

#Use this to build cstor-pool-mgmt
.PHONY: cstor-pool-mgmt
cstor-pool-mgmt:
	@echo "----------------------------"
	@echo "--> cstor-pool-mgmt           "
	@echo "----------------------------"
	@PNAME="cstor-pool-mgmt" CTLNAME=${POOL_MGMT} sh -c "'$(PWD)/buildscripts/build.sh'"

.PHONY: pool-mgmt-image
pool-mgmt-image: cstor-pool-mgmt
	@echo "----------------------------"
	@echo -n "--> cstor-pool-mgmt image "
	@echo "${HUB_USER}/${CSTOR_POOL_MGMT_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/cstor-pool-mgmt/${POOL_MGMT} buildscripts/cstor-pool-mgmt/
	@cd buildscripts/cstor-pool-mgmt && sudo docker build -t ${HUB_USER}/${CSTOR_POOL_MGMT_REPO_NAME}:${IMAGE_TAG} --build-arg BASE_IMAGE=${CSTOR_BASE_IMAGE} --build-arg BUILD_DATE=${BUILD_DATE} . --no-cache
	@rm buildscripts/cstor-pool-mgmt/${POOL_MGMT}

#Use this to build debug image of cstor-pool-mgmt
.PHONY: pool-mgmt-debug-image
pool-mgmt-debug-image:
	@echo "----------------------------"
	@echo -n "--> cstor-pool-mgmt debug image "
	@echo "${HUB_USER}/${CSTOR_POOL_MGMT_REPO_NAME}:inject"
	@echo "----------------------------"
	@PNAME=${CSP_OPERATOR_DEBUG} CTLNAME=${POOL_MGMT} BUILD_TAG="-tags=debug" sh -c "'$(PWD)/buildscripts/build.sh'"
	@cp bin/${CSP_OPERATOR_DEBUG}/${POOL_MGMT} buildscripts/${CSP_OPERATOR_DEBUG}/
	@cd buildscripts/${CSP_OPERATOR_DEBUG} && sudo docker build -t ${HUB_USER}/${CSTOR_POOL_MGMT_REPO_NAME}:inject --build-arg BASE_IMAGE=${CSTOR_BASE_IMAGE} --build-arg BUILD_DATE=${BUILD_DATE} . --no-cache
	@rm buildscripts/${CSP_OPERATOR_DEBUG}/${POOL_MGMT}

.PHONY: pool-mgmt-image.arm64
pool-mgmt-image.arm64: cstor-pool-mgmt
	@echo "----------------------------"
	@echo -n "--> cstor-pool-mgmt image "
	@echo "${HUB_USER}/${CSTOR_POOL_MGMT_REPO_NAME_ARM64}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/cstor-pool-mgmt/${POOL_MGMT} buildscripts/cstor-pool-mgmt/
	@cd buildscripts/cstor-pool-mgmt && sudo docker build -t ${HUB_USER}/${CSTOR_POOL_MGMT_REPO_NAME_ARM64}:${IMAGE_TAG} --build-arg BASE_IMAGE=${CSTOR_BASE_IMAGE_ARM64} --build-arg BUILD_DATE=${BUILD_DATE} . --no-cache
	@rm buildscripts/cstor-pool-mgmt/${POOL_MGMT}
