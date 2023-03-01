# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

ARG GO_VER=1.19
ARG ALPINE_VER=3.16

FROM golang:${GO_VER}-alpine${ALPINE_VER}

WORKDIR $GOPATH/src/github.com/hyperledger/fabric-contract-api-go
COPY . .
