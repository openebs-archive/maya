FROM alpine:3.10

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

ARG BUILD_DATE
LABEL org.label-schema.name="cvc-operator"
LABEL org.label-schema.description="Operator for OpenEBS cStor csi volumes"
LABEL org.label-schema.url="http://www.openebs.io/"
LABEL org.label-schema.vcs-url="https://github.com/openebs/maya"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$BUILD_DATE

ENTRYPOINT ["/usr/local/bin/cvc-operator"]
