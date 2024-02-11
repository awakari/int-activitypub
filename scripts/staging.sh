#!/bin/bash

export SLUG=ghcr.io/awakari/int-activitypub
export VERSION=latest
docker tag awakari/int-activitypub "${SLUG}":"${VERSION}"
docker push "${SLUG}":"${VERSION}"
