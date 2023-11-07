# Copyright 2020 The OpenEBS Authors. All rights reserved.
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
FROM golang:1.19.13 as build

ARG RELEASE_TAG
ARG BRANCH
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""

ENV GO111MODULE=on \
  GOOS=${TARGETOS} \
  GOARCH=${TARGETARCH} \
  GOARM=${TARGETVARIANT} \
  DEBIAN_FRONTEND=noninteractive \
  PATH="/root/go/bin:${PATH}" \
  RELEASE_TAG=${RELEASE_TAG} \
  BRANCH=${BRANCH}

WORKDIR /go/src/github.com/openebs/maya/

RUN apt-get update && apt-get install -y make git

COPY go.mod go.sum ./
# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download

COPY . .

RUN make mayactl apiserver
RUN chmod +x buildscripts/apiserver/entrypoint.sh

FROM openebs/linux-utils:2.12.x-ci
ENV MAYA_API_SERVER_NETWORK="eth0"

RUN apk add --no-cache \
    iproute2 \
    bash \
    curl \
    net-tools \
    mii-tool \
    procps \
    libc6-compat \
    ca-certificates

ARG DBUILD_DATE
ARG DBUILD_REPO_URL
ARG DBUILD_SITE_URL
LABEL org.label-schema.name="m-apiserver"
LABEL org.label-schema.description="API server for OpenEBS"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$DBUILD_DATE
LABEL org.label-schema.vcs-url=$DBUILD_REPO_URL
LABEL org.label-schema.url=$DBUILD_SITE_URL

RUN mkdir -p /etc/apiserver/orchprovider
RUN mkdir -p /etc/apiserver/specs

COPY --from=build /go/src/github.com/openebs/maya/buildscripts/apiserver/demo-vol1.yaml /etc/apiserver/specs/
COPY --from=build /go/src/github.com/openebs/maya/bin/apiserver/maya-apiserver /usr/local/bin/
COPY --from=build /go/src/github.com/openebs/maya/bin/maya/mayactl /usr/local/bin/
COPY --from=build /go/src/github.com/openebs/maya/buildscripts/apiserver/entrypoint.sh /usr/local/bin/

ENTRYPOINT entrypoint.sh "${MAYA_API_SERVER_NETWORK}"

EXPOSE 5656
