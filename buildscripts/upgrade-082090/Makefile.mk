# NOTE: The Upgrade design is changing and this file will be deprecated 
# in the future. Currently the upgrade process used by this only applies
# to 0.8.2 to 0.9.0. When the new implementation is updated to work for
# all upgrade paths including this one, this file will be removed. 

# This binay is not built as part of the CI, if there is an issue, 
# the image needs to be built and pushed manually. The image-tag
# needs to specifiy this is for 082090. 

# Specify the name for the binaries
UPGRADE-082090=upgrade-082090
UPGRADE-082090-TAG=082090-${IMAGE_TAG}

# build upgrade binary
.PHONY: upgrade-082090
upgrade-082090:
	@echo "----------------------------"
	@echo "--> ${UPGRADE-082090}       "
	@echo "----------------------------"
	@# PNAME is the sub-folder in ./bin where binary will be placed. 
	@# CTLNAME indicates the folder/pkg under cmd that needs to be built. 
	@# The output binary will be: ./bin/${PNAME}/<os-arch>/${CTLNAME}
	@# In this case as the binary is a sub directory under cmd, the binary
	@# will be: ./bin/upgrade-082090/<os-arch>/upgrade/upgrade-082090
	@# A copy of the binary will also be placed under: ./bin/upgrade-082090/
	@PNAME=${UPGRADE-082090} CTLNAME="upgrade/${UPGRADE-082090}" CGO_ENABLED=0 sh -c "'$(PWD)/buildscripts/build.sh'"

# build upgrade image
.PHONY: upgrade-image-082090
upgrade-image-082090: upgrade-082090
	@echo "-----------------------------------------------------------------------"
	@echo "--> ${UPGRADE-082090} image                                            "
	@echo "${IMAGE_ORG}/m-upgrade:${UPGRADE-082090-TAG}:${IMAGE_TAG}  "
	@echo "-----------------------------------------------------------------------"
	@# The binary is renamed as upgrade to keep it in sync with the upgrade image.
	@cp bin/${UPGRADE-082090}/${UPGRADE-082090} buildscripts/${UPGRADE-082090}/upgrade
	@cd buildscripts/${UPGRADE-082090} && \
	 sudo docker build --tag "${IMAGE_ORG}/m-upgrade:${UPGRADE-082090-TAG}" ${DBUILD_ARGS} .
	@rm buildscripts/${UPGRADE-082090}/upgrade

# cleanup upgrade build
.PHONY: cleanup-upgrade-082090
cleanup-upgrade-082090: 
	rm -rf ${GOPATH}/bin/${UPGRADE-082090}
