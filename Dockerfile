ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
LABEL maintainer="Kristof Keppens <kristof.keppens@ugent.be>"

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/pearl-exporter  /bin/pearl-exporter

EXPOSE      9115
ENTRYPOINT  [ "/bin/pearl-exporter" ]
