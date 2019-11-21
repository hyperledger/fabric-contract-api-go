#!/bin/bash

# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -e

if ! cat ./CHANGELOG.md | grep -q "## $1"; then
    echo "Changelog does not contain tag. Have you run ./.release/changelog.sh?"
    exit 1
fi