#!/usr/bin/env bash

#
# Run integration tests for silo.
#  - Requires docker
#

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT=${DIR}/../..
CONTAINER_NAME=silotest

echo "---> building"
set -eu
${ROOT}/docker/silo/build.sh
cd ${DIR}

echo "---> setting up"
set +e
unlink ${DIR}/silo/etc
ln -s ${ROOT}/docker/silo/dist ./silo/etc
ln -s ${ROOT}/cmd/silo/config.go ./silo/config.go
set -eu

PORT=`grep HttpPort silo/etc/silo.ini | head -n 1 | cut -d '=' -f 2`

tree  ./silo/

echo "---> booting silo"
echo ">" docker run -d --rm -p ${PORT}:${PORT} silo:test
docker run -d --name ${CONTAINER_NAME} --rm -p ${PORT}:${PORT} silo:test
set +e
echo "sleeping ..."
sleep 10
echo "... container logs (pre-test):"
docker ps --format "{{.ID}}" --filter "name=${CONTAINER_NAME}" | xargs docker logs

echo "---> starting test"
cd ./silo
go test

echo "---> cleaning up"
echo "container logs (post-test):"
docker ps --format "{{.ID}}" --filter "name=${CONTAINER_NAME}" | xargs docker logs

unlink ${DIR}/silo/etc
docker ps --format "{{.ID}}" --filter "name=${CONTAINER_NAME}" | xargs docker stop