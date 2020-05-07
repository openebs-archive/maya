
# Specify the name for the binaries
WEBHOOK=admission-server

# Specify the name of the docker repo for amd64
ADMISSION_SERVER_REPO_NAME?=admission-server

# Specify the name of the docker repo for arm64
ADMISSION_SERVER_REPO_NAME_ARM64?=admission-server-arm64

.PHONY: admission-server
admission-server:
	@echo "----------------------------"
	@echo -n "--> ${WEBHOOK} "
	@echo "----------------------------"
	@PNAME=${WEBHOOK} CTLNAME=${WEBHOOK} sh -c "'$(PWD)/buildscripts/build.sh'"

.PHONY: admission-server-image
admission-server-image: admission-server
	@echo "----------------------------"
	@echo -n "--> ${WEBHOOK} image"
	@echo "${IMAGE_ORG}/${ADMISSION_SERVER_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/${WEBHOOK}/${WEBHOOK} buildscripts/admission-server/
	@cd buildscripts/${WEBHOOK} && sudo docker build -t ${IMAGE_ORG}/${ADMISSION_SERVER_REPO_NAME}:${IMAGE_TAG} ${DBUILD_ARGS} .
	@rm buildscripts/${WEBHOOK}/${WEBHOOK}

.PHONY: admission-server-image.arm64
admission-server-image.arm64: admission-server
	@echo "----------------------------"
	@echo -n "--> ${WEBHOOK} image"
	@echo "${IMAGE_ORG}/${ADMISSION_SERVER_REPO_NAME_ARM64}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/${WEBHOOK}/${WEBHOOK} buildscripts/admission-server/
	@cd buildscripts/${WEBHOOK} && sudo docker build -t ${IMAGE_ORG}/${ADMISSION_SERVER_REPO_NAME_ARM64}:${IMAGE_TAG} ${DBUILD_ARGS} .
	@rm buildscripts/${WEBHOOK}/${WEBHOOK}

