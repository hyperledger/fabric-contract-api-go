#!/bin/bash

# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

fabric_dir="$(cd "$(dirname "$0")/../.." && pwd)"
source_dirs=$(go list -f  '{{ .Dir }}' ./... | sed s,"${fabric_dir}".,,g | cut -f 1 -d / | sort -u)

## Formatting
echo "running gofmt..."
gofmt_output="$(gofmt -l -s ${source_dirs})"
if [ -n "$gofmt_output" ]; then
    echo "The following files contain gofmt errors:"
    echo "$gofmt_output"
    echo "Please run 'gofmt -l -s -w' for these files."
    exit 1
fi

## Import management
echo "running goimports..."
goimports_output="$(goimports -l  ${source_dirs})"
if [ -n "$goimports_output" ]; then
    echo "The following files contain goimport errors:"
    echo "$goimports_output"
    echo "Please run 'goimports -l -w' for these files."
    exit 1
fi

## go vet
echo "running go vet..."
go vet ./...
