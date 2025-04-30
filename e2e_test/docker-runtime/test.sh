#!/bin/bash -e

docker run -e "EXEC_ID=$UID" -v $(pwd):/opt/code -v $(pwd)/e2e_test/docker-runtime/entrypoint.sh:/opt/entrypoint.sh -v $(pwd)/e2e_test/docker-runtime/entrypoint_run_tests.sh:/opt/entrypoint_run_tests.sh -w /opt --rm --entrypoint="/opt/entrypoint.sh" golang:1.23-bullseye