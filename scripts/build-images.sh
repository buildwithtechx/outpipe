#!/usr/bin/env bash

set -euo pipefail

tag="${TAG:-dev}"

docker build -f docker/Dockerfile.api -t "outpipe-api:$tag" .
docker build -f docker/Dockerfile.tunnel -t "outpipe-server:$tag" .
docker build -f docker/Dockerfile.cron -t "outpipe-cron:$tag" .
docker build -f docker/Dockerfile.check -t "outpipe-check:$tag" .
