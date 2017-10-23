#
# This Dockerfile builds a recent maya container with latest code
#

FROM alpine:3.6

RUN apk --no-cache add libc6-compat

COPY maya /usr/local/bin

ARG VERSION
ARG BUILD_DATE

LABEL org.label-schema.name="maya"
LABEL org.label-schema.description="CLI for OpenEBS"
LABEL org.label-schema.url="http://www.openebs.io/"
LABEL org.label-schema.vcs-url="https://github.com/openebs/maya"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.version=$VERSION
LABEL org.label-schema.build-date=$BUILD_DATE

ENTRYPOINT ["maya"]
