
# Specify the name for the binaries
UPGRADE=upgrade

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
	@echo "${HUB_USER}/${M_UPGRADE_REPO_NAME}:${IMAGE_TAG}"
	@echo "-----------------------------------------------"
	@cp bin/${UPGRADE}/${UPGRADE} buildscripts/${UPGRADE}/
	@cd buildscripts/${UPGRADE} && \
	 sudo docker build -t "${HUB_USER}/${M_UPGRADE_REPO_NAME}:${IMAGE_TAG}" --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/${UPGRADE}/${UPGRADE}

# cleanup upgrade build
.PHONY: cleanup-upgrade
cleanup-upgrade: 
	rm -rf ${GOPATH}/bin/${UPGRADE}
