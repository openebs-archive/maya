#openebs/cstor-base is the image that contains cstor related binaries and
#libraries - zpool, zfs, zrepl
#FROM openebs/cstor-base:ci
ARG BASE_IMAGE
FROM $BASE_IMAGE

ADD maya-exporter /usr/local/bin

ARG BUILD_DATE
LABEL org.label-schema.name="m-exporter"
LABEL org.label-schema.description="OpenEBS Maya Exporter"
LABEL org.label-schema.url="http://www.openebs.io/"
LABEL org.label-schema.vcs-url="https://github.com/openebs/maya"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$BUILD_DATE
CMD maya-exporter
EXPOSE 9500
