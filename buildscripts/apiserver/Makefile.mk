# Copyright © 2017 The OpenEBS Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Specify the name for the binaries
APISERVER=maya-apiserver

# Specify the name of the docker repo for amd64
M_APISERVER_REPO_NAME?=m-apiserver

# Specify the name of the docker repo for arm64
M_APISERVER_REPO_NAME_ARM64?=m-apiserver-arm64

# Use this to build only the maya apiserver.
.PHONY: apiserver
apiserver:
	@echo "----------------------------"
	@echo "--> maya-apiserver               "
	@echo "----------------------------"
	@PNAME="apiserver" CTLNAME=${APISERVER} sh -c "'$(PWD)/buildscripts/build.sh'"

#TODO: 
# Currently both mayactl & apiserver binaries are pushed into
# m-apiserver image. This is going to be decoupled soon.
.PHONY: apiserver-image
apiserver-image: apiserver
	@echo "----------------------------"
	@echo -n "--> apiserver image "
	@echo "${IMAGE_ORG}/${M_APISERVER_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/apiserver/${APISERVER} buildscripts/apiserver/
	#@cp bin/maya/${MAYACTL} buildscripts/apiserver/
	@cd buildscripts/apiserver && sudo docker build -t ${IMAGE_ORG}/${M_APISERVER_REPO_NAME}:${IMAGE_TAG} ${DBUILD_ARGS} .
	@rm buildscripts/apiserver/${APISERVER}
	#@rm buildscripts/apiserver/${MAYACTL}
#TODO: remove here too
.PHONY: rhel-apiserver-image
rhel-apiserver-image: mayactl apiserver
	@echo "----------------------------"
	@echo -n "--> rhel based apiserver image "
	@echo "${IMAGE_ORG}/${M_APISERVER_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/apiserver/${APISERVER} buildscripts/apiserver/
	# @cp bin/maya/${MAYACTL} buildscripts/apiserver/
	@cd buildscripts/apiserver && sudo docker build -t ${IMAGE_ORG}/${M_APISERVER_REPO_NAME}:${IMAGE_TAG} -f Dockerfile.rhel --build-arg VERSION=${VERSION} .
	@rm buildscripts/apiserver/${APISERVER}
	# @rm buildscripts/apiserver/${MAYACTL}
#TODO: remove here too 
.PHONY: apiserver-image.arm64
apiserver-image.arm64: mayactl apiserver
	@echo "----------------------------"
	@echo -n "--> apiserver image "
	@echo "${IMAGE_ORG}/${M_APISERVER_REPO_NAME_ARM64}:${IMAGE_TAG}"
	@echo "----------------------------"
	@cp bin/apiserver/${APISERVER} buildscripts/apiserver/
	# @cp bin/maya/${MAYACTL} buildscripts/apiserver/
	@cd buildscripts/apiserver && sudo docker build -t ${IMAGE_ORG}/${M_APISERVER_REPO_NAME_ARM64}:${IMAGE_TAG} -f Dockerfile.arm64 ${DBUILD_ARGS} --build-arg BASE_IMAGE=${BASE_DOCKER_IMAGE_ARM64} .
	@rm buildscripts/apiserver/${APISERVER}
	# @rm buildscripts/apiserver/${MAYACTL}

