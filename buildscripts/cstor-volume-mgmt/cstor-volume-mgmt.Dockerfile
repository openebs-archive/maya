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
ARG BASE_IMAGE
FROM golang:1.14.7 as build

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

RUN make cstor-volume-mgmt
RUN chmod +x buildscripts/cstor-volume-mgmt/entrypoint.sh

FROM ubuntu:16.04
RUN apt-get update; exit 0
RUN apt-get -y install rsyslog

ARG DBUILD_DATE
ARG DBUILD_REPO_URL
ARG DBUILD_SITE_URL
LABEL org.label-schema.name="cstor-volume-mgmt"
LABEL org.label-schema.description="OpenEBS cStor Volume Operator"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$DBUILD_DATE
LABEL org.label-schema.vcs-url=$DBUILD_REPO_URL
LABEL org.label-schema.url=$DBUILD_SITE_URL

RUN mkdir -p /usr/local/etc/istgt

COPY --from=build /go/src/github.com/openebs/maya/bin/cstor-volume-mgmt/cstor-volume-mgmt /usr/local/bin/
COPY --from=build /go/src/github.com/openebs/maya/buildscripts/cstor-volume-mgmt/entrypoint.sh /usr/local/bin/

ENTRYPOINT entrypoint.sh
EXPOSE 7676 7777