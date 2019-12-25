
# Specify the name for the binaries
VOLUME_MGMT=cstor-volume-mgmt

# Specify the name of the docker repo for amd64
CSTOR_VOLUME_MGMT_REPO_NAME?=cstor-volume-mgmt

# Specify the name of the docker repo for arm64
CSTOR_VOLUME_MGMT_REPO_NAME_ARM64?=cstor-volume-mgmt-arm64

#Use this to build cstor-volume-mgmt
.PHONY: cstor-volume-mgmt
cstor-volume-mgmt:
	@echo "----------------------------"
	@echo "--> cstor-volume-mgmt           "
	@echo "----------------------------"
	@PNAME="cstor-volume-mgmt" CTLNAME=${VOLUME_MGMT} sh -c "'$(PWD)/buildscripts/build.sh'"

.PHONY: protobuf
protobuf:
	@echo "----------------------------"
	@echo "--> protobuf           "
	@echo "----------------------------"
	@protoc -I $(PWD)/pkg/apis/openebs.io/v1alpha1/ \
    -I${GOPATH}/src \
    --go_out=plugins=grpc:$(PWD)/pkg/client/generated/cstor-volume-mgmt/v1alpha1 \
    $(PWD)/pkg/apis/openebs.io/v1alpha1/cstorvolume.proto

.PHONY: volume-mgmt-image
volume-mgmt-image: cstor-volume-mgmt
	@echo "----------------------------"
	@echo -n "--> cstor-volume-mgmt image "
	@echo "${HUB_USER}/${CSTOR_VOLUME_MGMT_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/cstor-volume-mgmt/${VOLUME_MGMT} buildscripts/cstor-volume-mgmt/
	@cd buildscripts/cstor-volume-mgmt && sudo docker build -t ${HUB_USER}/${CSTOR_VOLUME_MGMT_REPO_NAME}:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/cstor-volume-mgmt/${VOLUME_MGMT}

.PHONY: volume-mgmt-image.arm64
volume-mgmt-image.arm64: cstor-volume-mgmt
	@echo "----------------------------"
	@echo -n "--> cstor-volume-mgmt image "
	@echo "${HUB_USER}/${CSTOR_VOLUME_MGMT_REPO_NAME_ARM64}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/cstor-volume-mgmt/${VOLUME_MGMT} buildscripts/cstor-volume-mgmt/
	@cd buildscripts/cstor-volume-mgmt && sudo docker build -t ${HUB_USER}/${CSTOR_VOLUME_MGMT_REPO_NAME_ARM64}:${IMAGE_TAG} -f Dockerfile.arm64 --build-arg BUILD_DATE=${BUILD_DATE} --build-arg BASE_IMAGE=${BASE_DOCKER_IMAGE_ARM64} .
	@rm buildscripts/cstor-volume-mgmt/${VOLUME_MGMT}
