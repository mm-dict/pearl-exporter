# Needs to be defined before including Makefile.common to auto-generate targets
DOCKER_ARCHS ?= amd64
DOCKER_REPO             ?= quay.io/mm-dict

include Makefile.common

DOCKER_IMAGE_NAME       ?= pearl-exporter
