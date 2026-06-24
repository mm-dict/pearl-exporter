#!/usr/bin/env bash
# Build and push a multi-arch (amd64 + arm64) image for pearl-exporter.
set -euo pipefail

IMAGE="${IMAGE:-kristofkeppens/pearl-exporter}"
TAG="${TAG:-latest}"
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"
BUILDER="${BUILDER:-pearl-exporter-builder}"

cd "$(dirname "$0")"

# Ensure a buildx builder with multi-arch support exists.
if ! docker buildx inspect "$BUILDER" >/dev/null 2>&1; then
  docker buildx create --name "$BUILDER" --driver docker-container --use
else
  docker buildx use "$BUILDER"
fi

docker buildx build \
  --platform "$PLATFORMS" \
  --tag "$IMAGE:$TAG" \
  --push \
  .
