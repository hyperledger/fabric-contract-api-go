#!/bin/bash

# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

## Formatting
echo "running gofmt..."
gofmt_output="$(gofmt -l -s . | grep -v .azure-pipelines && exit 1 || exit 0)"
if [ -n "$gofmt_output" ]; then
    echo "The following files contain gofmt errors:"
    echo "$gofmt_output"
    echo "Please run 'gofmt -l -s -w' for these files."
    exit 1
fi

## Import management
echo "running goimports..."
goimports_output="$(goimports -l  . | grep -v .azure-pipelines && exit 1 || exit 0)"
if [ -n "$goimports_output" ]; then
    echo "The following files contain goimport errors:"
    echo "$goimports_output"
    echo "Please run 'goimports -l -w' for these files."
    exit 1
fi

## go vet
echo "running go vet..."
go vet ./...

## golint
echo "running golint..."
golint -set_exit_status ./...
