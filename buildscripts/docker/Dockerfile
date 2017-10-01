#
# This Dockerfile builds a recent maya container with latest code
#

FROM ubuntu:16.04

COPY maya /usr/local/bin

ARG VERSION
ARG BUILD_DATE

LABEL org.label-schema.vendor="OpenEBS"
LABEL org.label-schema.version=$VERSION
LABEL org.label-schema.build-date=$BUILD_DATE

ENTRYPOINT ["maya"]
