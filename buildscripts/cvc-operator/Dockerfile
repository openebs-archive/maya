FROM alpine:3.11.5

RUN apk add --no-cache \
    iproute2 \
    bash \
    curl \
    net-tools \
    mii-tool \
    procps \
    libc6-compat \
    ca-certificates

COPY cvc-operator /usr/local/bin/cvc-operator

ARG ARCH
ARG DBUILD_DATE
ARG DBUILD_REPO_URL
ARG DBUILD_SITE_URL
LABEL org.label-schema.name="cvc-operator"
LABEL org.label-schema.description="Operator for OpenEBS cStor csi volumes"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$DBUILD_DATE
LABEL org.label-schema.vcs-url=$DBUILD_REPO_URL
LABEL org.label-schema.url=$DBUILD_SITE_URL

ENTRYPOINT ["/usr/local/bin/cvc-operator"]
