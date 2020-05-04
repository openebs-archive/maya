
# Specify the name for binaries
CSPI_MGMT=cspi-mgmt

# Specify the name of the docker repo for amd64
CSPI_MGMT_REPO_NAME?=cspi-mgmt

# Specify the name of the docker repo for arm64
CSPI_MGMT_REPO_NAME_ARM64?=cspi-mgmt-arm64

#Use this to build cspi-mgmt
.PHONY: cspi-mgmt
cspi-mgmt:
	@echo "----------------------------"
	@echo "--> cspi-mgmt           "
	@echo "----------------------------"
	@PNAME="cspi-mgmt" CTLNAME=${CSPI_MGMT} sh -c "'$(PWD)/buildscripts/build.sh'"

.PHONY: cspi-mgmt-image
cspi-mgmt-image: cspi-mgmt
	@echo "----------------------------"
	@echo -n "--> cspi-mgmt image "
	@echo "${IMAGE_ORG}/${CSPI_MGMT_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/cspi-mgmt/${CSPI_MGMT} buildscripts/cspi-mgmt/
	@cd buildscripts/cspi-mgmt && sudo docker build -t ${IMAGE_ORG}/${CSPI_MGMT_REPO_NAME}:${IMAGE_TAG} --build-arg BASE_IMAGE=${CSTOR_BASE_IMAGE} ${DBUILD_ARGS} . --no-cache
	@rm buildscripts/cspi-mgmt/${CSPI_MGMT}

.PHONY: cspi-mgmt-image.arm64
cspi-mgmt-image.arm64: cspi-mgmt
	@echo "----------------------------"
	@echo -n "--> cspi-mgmt image "
	@echo "${IMAGE_ORG}/${CSPI_MGMT_REPO_NAME_ARM64}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/cspi-mgmt/${CSPI_MGMT} buildscripts/cspi-mgmt/
	@cd buildscripts/cspi-mgmt && sudo docker build -t ${IMAGE_ORG}/${CSPI_MGMT_REPO_NAME_ARM64}:${IMAGE_TAG} --build-arg BASE_IMAGE=${CSTOR_BASE_IMAGE_ARM64} ${DBUILD_ARGS} . --no-cache
	@rm buildscripts/cspi-mgmt/${CSPI_MGMT}
