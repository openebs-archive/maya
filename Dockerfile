#
# This Dockerfile builds the maya components using the Makefile
# 

FROM golang:latest

ARG BUILD_DATE

LABEL org.label-schema.name="maya"
LABEL org.label-schema.description="OpenEBS Storage Orchestration Engine"
LABEL org.label-schema.url="http://www.openebs.io/"
LABEL org.label-schema.vcs-url="https://github.com/openebs/maya"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$BUILD_DATE

# Setup environment
ENV PWD=/usr/local/go/src/github.com/openebs/maya

RUN apt-get update && \
    apt-get install -y zip

WORKDIR /usr/local/go/src/github.com/openebs/maya

# TODO: Add entrypoint to improve building
COPY . .

RUN make bootstrap && make bin && make apiserver && make maya-agent && make install