# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

ARG DEBIAN_BASE=buster
ARG GO_VER=1.13.4

FROM golang:${GO_VER}-${DEBIAN_BASE} as golang
ADD tools /tools
WORKDIR tools
ENV GO111MODULE on

RUN GOBIN=/usr/local/bin go install golang.org/x/lint/golint
RUN GOBIN=/usr/local/bin go install golang.org/x/tools/cmd/goimports

FROM golang:${GO_VER}-${DEBIAN_BASE}
COPY --from=golang /usr/local/bin/* /usr/local/bin/
