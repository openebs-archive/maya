#
# This builds openebs cluster image using 
# its latest binary
#

FROM alpine:3.6

# copy the latest binary
COPY openebs-cluster /usr/local/bin/

ARG BUILD_DATE

LABEL org.label-schema.name="openebs-cluster"
LABEL org.label-schema.description="operator for openebs"
LABEL org.label-schema.url="http://www.openebs.io/"
LABEL org.label-schema.vcs-url="https://github.com/openebs/maya"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$BUILD_DATE

ENTRYPOINT ["openebs-cluster"]
