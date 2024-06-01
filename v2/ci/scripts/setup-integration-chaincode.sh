#!/bin/bash

# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Adds to go.mod file of each of the integration chaincodes a pointer to use the local files instead of github
# for the contract api go packages. Runs go mod vendor to put these files into a vendor folder so they are used 
# by the integration tool which doesn't have the local files. Removes the replace to stop the integration tests
# breaking as the dot path doesn't exist there.

# Note the actually committed go.mod files of the chaincode already have the 
# 'replace github.com/hyperledger/fabric-contract-api-go => ../../..' I

set -e -u -o pipefail
ROOTDIR=$(cd "$(dirname "$0")" && pwd)

CHAINCODE_DIR=$ROOTDIR/../../integrationtest/chaincode
ls -lart $CHAINCODE_DIR
pushd $CHAINCODE_DIR
for testCC in */; do
    pushd $testCC
    go mod vendor 
    popd
done
popd

