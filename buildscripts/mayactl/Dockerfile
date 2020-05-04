#
# This Dockerfile builds a recent maya container with latest code
#

FROM alpine:3.11.5

RUN apk --no-cache add libc6-compat

COPY maya /usr/local/bin

ARG VERSION

ARG ARCH
ARG DBUILD_DATE
ARG DBUILD_REPO_URL
ARG DBUILD_SITE_URL
LABEL org.label-schema.name="maya"
LABEL org.label-schema.description="CLI for OpenEBS"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$DBUILD_DATE
LABEL org.label-schema.vcs-url=$DBUILD_REPO_URL
LABEL org.label-schema.url=$DBUILD_SITE_URL

ENTRYPOINT ["maya"]
