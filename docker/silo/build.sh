#!/usr/bin/env bash

#
# Builds example silo docker image
# - creates ssl certs at random
# - generates config / docker build
# - builds image (using local repository)
# - saves config & ssl files to /dist/ folder
#
# Obviously, this requires docker.
#

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT=${DIR}/../../
TOOL_DIR=${ROOT}/tools/

export SILO_HOST=0.0.0.0
export SILO_PORT=9150
export SILO_SSL_CERT=ssl.cert
export SILO_SSL_KEY=ssl.key
export SILO_MAX_DATA_BYTES=1000000
export SILO_MAX_KEY_BYTES=100
export SILO_ENCRYPTION_KEY=wellthisreallyshouldbechangedtosomethingelseiguess
export SILO_STORE_LOCATION=/tmp/silo
export SILO_ROLE_PASS_READ=readpassword
export SILO_ROLE_PASS_READWRITE=readwritepassword
export SILO_ROLE_PASS_ALL=allpassword

export DOCKER_GOPATH="\${GOPATH}"

echo "==== Preparing"
set -eu
cd ${DIR}
set +e
rm -vrf build
rm -vrf dist
set -eu

# generate ssl certificates
echo "==== Generating test ssl certs"
${TOOL_DIR}/generate_ssl.sh

# inject config vars
echo "==== Writing config & dockerfiles"
envsubst < silo.ini.template > silo.ini
envsubst < Dockerfile.template > Dockerfile

# copy local code into temp dir
echo "==== Creating build dir"
mkdir build
mkdir dist
cp -v silo.ini dist/
mv -v silo.ini build/
cp -v ${TOOL_DIR}/ssl.* dist/
mv -v ${TOOL_DIR}/ssl.* build/
cp -v ${ROOT}*.go build/
cp -vr ${ROOT}/cmd build/

# print state of build dir
set +e
tree dist
tree build

echo "==== Building"
docker build -t silo:test ./

echo "==== Cleaning up"
rm -vfr build
rm -v Dockerfile

echo "==== Done"
