# Copyright 2018-2020 The OpenEBS Authors. All rights reserved.
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

# Build cstor-operator docker images with buildx
# Experimental docker feature to build cross platform multi-architecture docker images
# https://docs.docker.com/buildx/working-with-buildx/

# ==============================================================================
# Build Options

export DBUILD_ARGS=--build-arg DBUILD_DATE=${DBUILD_DATE} --build-arg DBUILD_REPO_URL=${DBUILD_REPO_URL} --build-arg DBUILD_SITE_URL=${DBUILD_SITE_URL} --build-arg RELEASE_TAG=${RELEASE_TAG}

ifeq (${TAG}, )
  export TAG=ci
endif

CSTOR_BASE_IMAGE= ${IMAGE_ORG}/cstor-base:${TAG}

# default list of platforms for which multiarch image is built
ifeq (${PLATFORMS}, )
	export PLATFORMS="linux/amd64,linux/arm64"
endif

# if IMG_RESULT is unspecified, by default the image will be pushed to registry
ifeq (${IMG_RESULT}, load)
	export PUSH_ARG="--load"
    # if load is specified, image will be built only for the build machine architecture.
    export PLATFORMS="local"
else ifeq (${IMG_RESULT}, cache)
	# if cache is specified, image will only be available in the build cache, it won't be pushed or loaded
	# therefore no PUSH_ARG will be specified
else
	export PUSH_ARG="--push"
endif

# Name of the multiarch image for m-apiserver
DOCKERX_IMAGE_APISERVER:=${IMAGE_ORG}/m-apiserver:${TAG}

# Name of the multiarch image for cstor-pool-mgmt
DOCKERX_IMAGE_CSTOR_POOL_MGMT:=${IMAGE_ORG}/cstor-pool-mgmt:${TAG}

# Name of the multiarch image for cstor-volume-mgmt
DOCKERX_IMAGE_CSTOR_VOLUME_MGMT:=${IMAGE_ORG}/cstor-volume-mgmt:${TAG}

# Name of the multiarch image for admission-server
DOCKERX_IMAGE_ADMISSION_SERVER:=${IMAGE_ORG}/admission-server:${TAG}

# Name of the multiarch image for m-upgrade
DOCKERX_IMAGE_UPGRADE:=${IMAGE_ORG}/m-upgrade:${TAG}

.PHONY: docker.buildx
docker.buildx:
	export DOCKER_CLI_EXPERIMENTAL=enabled
	@if ! docker buildx ls | grep -q container-builder; then\
		docker buildx create --platform ${PLATFORMS} --name container-builder --use;\
	fi
	@docker buildx build --platform "${PLATFORMS}" \
		-t "$(DOCKERX_IMAGE_NAME)" ${BUILD_ARGS} \
		-f $(PWD)/buildscripts/$(COMPONENT)/$(COMPONENT).Dockerfile \
		. ${PUSH_ARG}
	@echo "--> Build docker image: $(DOCKERX_IMAGE_NAME)"
	@echo

.PHONY: docker.buildx.apiserver
docker.buildx.apiserver: DOCKERX_IMAGE_NAME=$(DOCKERX_IMAGE_APISERVER)
docker.buildx.apiserver: COMPONENT=apiserver
docker.buildx.apiserver: BUILD_ARGS=$(DBUILD_ARGS)
docker.buildx.apiserver: docker.buildx

.PHONY: docker.buildx.cstor-volume-mgmt
docker.buildx.cstor-volume-mgmt: DOCKERX_IMAGE_NAME=$(DOCKERX_IMAGE_CSTOR_VOLUME_MGMT)
docker.buildx.cstor-volume-mgmt: COMPONENT=cstor-volume-mgmt
docker.buildx.cstor-volume-mgmt: BUILD_ARGS=$(DBUILD_ARGS)
docker.buildx.cstor-volume-mgmt: docker.buildx

.PHONY: docker.buildx.cstor-pool-mgmt
docker.buildx.cstor-pool-mgmt: DOCKERX_IMAGE_NAME=$(DOCKERX_IMAGE_CSTOR_POOL_MGMT)
docker.buildx.cstor-pool-mgmt: COMPONENT=cstor-pool-mgmt
docker.buildx.cstor-pool-mgmt: BUILD_ARGS=--build-arg BASE_IMAGE=$(CSTOR_BASE_IMAGE) ${DBUILD_ARGS}
docker.buildx.cstor-pool-mgmt: docker.buildx

.PHONY: docker.buildx.admission-server
docker.buildx.admission-server: DOCKERX_IMAGE_NAME=$(DOCKERX_IMAGE_ADMISSION_SERVER)
docker.buildx.admission-server: COMPONENT=admission-server
docker.buildx.admission-server: BUILD_ARGS=$(DBUILD_ARGS)
docker.buildx.admission-server: docker.buildx

.PHONY: docker.buildx.upgrade
docker.buildx.upgrade: DOCKERX_IMAGE_NAME=$(DOCKERX_IMAGE_UPGRADE)
docker.buildx.upgrade: COMPONENT=upgrade
docker.buildx.upgrade: BUILD_ARGS=$(DBUILD_ARGS)
docker.buildx.upgrade: docker.buildx

.PHONY: buildx.push.apiserver
buildx.push.apiserver:
	BUILDX=true DIMAGE=${IMAGE_ORG}/m-apiserver ./buildscripts/buildxpush.sh

.PHONY: buildx.push.cstor-volume-mgmt
buildx.push.cstor-volume-mgmt:
	BUILDX=true DIMAGE=${IMAGE_ORG}/cstor-volume-mgmt ./buildscripts/buildxpush.sh

.PHONY: buildx.push.cstor-pool-mgmt
buildx.push.cstor-pool-mgmt:
	BUILDX=true DIMAGE=${IMAGE_ORG}/cstor-pool-mgmt ./buildscripts/buildxpush.sh

.PHONY: buildx.push.admission-server
buildx.push.admission-server:
	BUILDX=true DIMAGE=${IMAGE_ORG}/admission-server ./buildscripts/buildxpush.sh

.PHONY: buildx.push.upgrade
buildx.push.upgrade:
	BUILDX=true DIMAGE=${IMAGE_ORG}/m-upgrade ./buildscripts/buildxpush.sh
