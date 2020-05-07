
# Specify the name for the binaries
UPGRADE=upgrade

# Specify the name of the docker repo for amd64
UPGRADE_REPO_NAME="m-upgrade"

# Specify the name of the docker repo for arm64
UPGRADE_REPO_NAME_ARM64="m-upgrade-arm64"

# build upgrade binary
.PHONY: upgrade
upgrade:
	@echo "----------------------------"
	@echo "--> ${UPGRADE}              "
	@echo "----------------------------"
	@# PNAME is the sub-folder in ./bin where binary will be placed. 
	@# CTLNAME indicates the folder/pkg under cmd that needs to be built. 
	@# The output binary will be: ./bin/${PNAME}/<os-arch>/${CTLNAME}
	@# A copy of the binary will also be placed under: ./bin/${PNAME}/${CTLNAME}
	@PNAME=${UPGRADE} CTLNAME=${UPGRADE} CGO_ENABLED=0 sh -c "'$(PWD)/buildscripts/build.sh'"

# build upgrade image
.PHONY: upgrade-image
upgrade-image: upgrade
	@echo "-----------------------------------------------"
	@echo "--> ${UPGRADE} image                           "
	@echo "${IMAGE_ORG}/${UPGRADE_REPO_NAME}:${IMAGE_TAG}"
	@echo "-----------------------------------------------"
	@cp bin/${UPGRADE}/${UPGRADE} buildscripts/${UPGRADE}/
	@cd buildscripts/${UPGRADE} && \
	 sudo docker build -t "${IMAGE_ORG}/${UPGRADE_REPO_NAME}:${IMAGE_TAG}" ${DBUILD_ARGS} .
	@rm buildscripts/${UPGRADE}/${UPGRADE}

.PHONY: upgrade-image.arm64
upgrade-image.arm64: upgrade
	@echo "-----------------------------------------------"
	@echo "--> ${UPGRADE} image                           "
	@echo "${IMAGE_ORG}/${UPGRADE_REPO_NAME_ARM64}:${IMAGE_TAG}"
	@echo "-----------------------------------------------"
	@cp bin/${UPGRADE}/${UPGRADE} buildscripts/${UPGRADE}/
	@cd buildscripts/${UPGRADE} && \
	 sudo docker build -t "${IMAGE_ORG}/${UPGRADE_REPO_NAME_ARM64}:${IMAGE_TAG}" ${DBUILD_ARGS} .
	@rm buildscripts/${UPGRADE}/${UPGRADE}




# cleanup upgrade build
.PHONY: cleanup-upgrade
cleanup-upgrade: 
	rm -rf ${GOPATH}/bin/${UPGRADE}
