#
# This builds openebs cluster image using 
# its latest binary
#

FROM alpine:3.11.5

# copy the latest binary
COPY openebs-cluster /usr/local/bin/

ARG ARCH
ARG DBUILD_DATE
ARG DBUILD_REPO_URL
ARG DBUILD_SITE_URL

LABEL org.label-schema.name="openebs-cluster"
LABEL org.label-schema.description="operator for openebs"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$DBUILD_DATE
LABEL org.label-schema.vcs-url=$DBUILD_REPO_URL
LABEL org.label-schema.url=$DBUILD_SITE_URL

ENTRYPOINT ["openebs-cluster"]
