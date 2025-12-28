#!/usr/bin/env bash
set -euo pipefail

# Usage: ./publish.sh --build --publish
#
# Environment overrides:
# - DOCKERHUB_USERNAME / DOCKERHUB_PASSWORD: if set, script will perform `docker login`

MMR_VERSION=$(git describe --tags)
DOCKERHUB_REPO=watermelon0339/matrix-media-repo

LOCAL_IMAGE="matrix-media-repo:${MMR_VERSION}"
REMOTE_IMAGE="${DOCKERHUB_REPO}:${MMR_VERSION}"

# Parse args
BUILD=false
PUBLISH=false
for arg in "$@"; do
    case "$arg" in
        --build) BUILD=true ;;
        --publish) PUBLISH=true ;;
        --help|-h)
            echo "Usage: ./publish.sh [--build] [--publish]"
            exit 0
            ;;
        *)
            echo "Unknown option: $arg"
            exit 1
            ;;
    esac
done

if [ "$BUILD" = false ] && [ "$PUBLISH" = false ]; then
    echo "Nothing to do. Use --build and/or --publish."
    exit 0
fi

if ! $BUILD; then
    echo "Skipping build step."
else
    echo "Building image: ${LOCAL_IMAGE} (base MMR_VERSION=${MMR_VERSION})"
    docker build -t "${LOCAL_IMAGE}" .
    echo "Tagging ${LOCAL_IMAGE} -> ${REMOTE_IMAGE}"
    docker tag "${LOCAL_IMAGE}" "${REMOTE_IMAGE}"
fi

if ! $PUBLISH; then
    echo "Skipping publish step."
    exit 0
fi

if [[ -n "${DOCKERHUB_USERNAME:-}" && -n "${DOCKERHUB_PASSWORD:-}" ]]; then
	echo "Logging into Docker Hub as ${DOCKERHUB_USERNAME}"
	echo "${DOCKERHUB_PASSWORD}" | docker login --username "${DOCKERHUB_USERNAME}" --password-stdin
else
	echo "Skipping docker login (set DOCKERHUB_USERNAME and DOCKERHUB_PASSWORD to enable)."
	echo "Ensure you are already logged in: docker login"
fi

echo "Pushing ${REMOTE_IMAGE}"
docker push "${REMOTE_IMAGE}"
