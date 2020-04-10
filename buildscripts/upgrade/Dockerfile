#
# This builds openebs upgrade image using 
# its latest binary
#

FROM alpine:3.11.5

# copy the latest binary
COPY upgrade /usr/local/bin/upgrade

ARG BUILD_DATE

LABEL org.label-schema.name="upgrade"
LABEL org.label-schema.description="upgrades openebs components"
LABEL org.label-schema.url="https://openebs.io/"
LABEL org.label-schema.vcs-url="https://github.com/openebs/maya"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$BUILD_DATE

ENTRYPOINT ["upgrade"]
