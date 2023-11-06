#!/bin/bash -e

docker run -v $(pwd):/opt/code -v $(pwd)/e2e_test/docker-runtime/entrypoint.sh:/opt/entrypoint.sh -w /opt --rm --entrypoint="/opt/entrypoint.sh" golang:1.20-bullseye