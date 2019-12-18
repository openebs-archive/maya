
# Specify the name for the binaries
MIGRATE=migrate

# build migrate binary
.PHONY: migrate
migrate:
	@echo "----------------------------"
	@echo "--> ${MIGRATE}              "
	@echo "----------------------------"
	@# PNAME is the sub-folder in ./bin where binary will be placed. 
	@# CTLNAME indicates the folder/pkg under cmd that needs to be built. 
	@# The output binary will be: ./bin/${PNAME}/<os-arch>/${CTLNAME}
	@# A copy of the binary will also be placed under: ./bin/${PNAME}/${CTLNAME}
	@PNAME=${MIGRATE} CTLNAME=${MIGRATE} CGO_ENABLED=0 sh -c "'$(PWD)/buildscripts/build.sh'"

# build migrate image
.PHONY: migrate-image
migrate-image: migrate
	@echo "-----------------------------------------------"
	@echo "--> ${MIGRATE} image                           "
	@echo "${HUB_USER}/${M_MIGRATE_REPO_NAME}:${IMAGE_TAG}"
	@echo "-----------------------------------------------"
	@cp bin/${MIGRATE}/${MIGRATE} buildscripts/${MIGRATE}/
	@cd buildscripts/${MIGRATE} && \
	 sudo docker build -t "${HUB_USER}/${M_MIGRATE_REPO_NAME}:${IMAGE_TAG}" --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/${MIGRATE}/${MIGRATE}

# cleanup migrate build
.PHONY: cleanup-migrate
cleanup-migrate: 
	rm -rf ${GOPATH}/bin/${MIGRATE}
