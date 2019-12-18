#
# This builds openebs migratte image using 
# its latest binary
#

FROM alpine:3.9

# copy the latest binary
COPY migrate /usr/local/bin/migrate

ARG BUILD_DATE

LABEL org.label-schema.name="migrate"
LABEL org.label-schema.description="migrates openebs components"
LABEL org.label-schema.url="https://openebs.io/"
LABEL org.label-schema.vcs-url="https://github.com/openebs/maya"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$BUILD_DATE

ENTRYPOINT ["migrate"]
