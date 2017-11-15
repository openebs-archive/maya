FROM alpine:3.6
ADD maya-volume-exporter /usr/local/bin
RUN apk add --no-cache \
     iproute2 \
     curl \
     net-tools \
     mii-tool \
     procps \
     libc6-compat

ARG BUILD_DATE
LABEL org.label-schema.name="m-exporter"
LABEL org.label-schema.description="OpenEBS Volume Exporter"
LABEL org.label-schema.url="http://www.openebs.io/"
LABEL org.label-schema.vcs-url="https://github.com/openebs/maya"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$BUILD_DATE
CMD maya-volume-exporter
EXPOSE 9500
