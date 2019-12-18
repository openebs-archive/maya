
# Specify the name for the binaries
EXPORTER=maya-exporter

# Specify the name of docker repo for amd64
M_EXPORTER_REPO_NAME?=m-exporter

# Specify the name of docker repo for arm64
M_EXPORTER_REPO_NAME_ARM64?=m-exporter-arm64


# Use this to build only the maya-exporter.
.PHONY: exporter
exporter:
	@echo "----------------------------"
	@echo "--> maya-exporter              "
	@echo "----------------------------"
	@PNAME="exporter" CTLNAME=${EXPORTER} sh -c "'$(PWD)/buildscripts/build.sh'"

# m-exporter image. This is going to be decoupled soon.
.PHONY: exporter-image
exporter-image: exporter
	@echo "----------------------------"
	@echo -n "--> m-exporter image "
	@echo "${HUB_USER}/${M_EXPORTER_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/exporter/${EXPORTER} buildscripts/exporter/
	@cd buildscripts/exporter && sudo docker build -t ${HUB_USER}/${M_EXPORTER_REPO_NAME}:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} --build-arg BASE_IMAGE=${CSTOR_BASE_IMAGE} .
	@rm buildscripts/exporter/${EXPORTER}

.PHONY: exporter-image.arm64
exporter-image.arm64: exporter
	@echo "----------------------------"
	@echo -n "--> m-exporter image "
	@echo "${HUB_USER}/${M_EXPORTER_REPO_NAME_ARM64}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/exporter/${EXPORTER} buildscripts/exporter/
	@cd buildscripts/exporter && sudo docker build -t ${HUB_USER}/${M_EXPORTER_REPO_NAME_ARM64}:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} --build-arg BASE_IMAGE=${CSTOR_BASE_IMAGE_ARM64} .
	@rm buildscripts/exporter/${EXPORTER}
