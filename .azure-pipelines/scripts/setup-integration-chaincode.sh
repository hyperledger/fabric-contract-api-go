#!/bin/bash

# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Adds to go.mod file of each of the integration chaincodes a pointer to use the local files instead of github
# for the contract api go packages. Runs go mod vendor to put these files into a vendor folder so they are used 
# by the integration tool which doesn't have the local files. Removes the replace to stop the integration tests
# breaking as the dot path doesn't exist there.

DIR=$(pwd)
CHAINCODE_DIR=$DIR/.azure-pipelines/resources/chaincode

cd $CHAINCODE_DIR

for testCC in */; do
    cd $testCC
    cat go.mod > tmp.go.mod
    echo 'replace github.com/hyperledger/fabric-contract-api-go => ../../../..' >> go.mod
    go mod vendor
    
    cat tmp.go.mod > go.mod

    cd $CHAINCODE_DIR
done

cd $DIR
