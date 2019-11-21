# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

ARG DEBIAN_BASE=buster
ARG GO_VER=1.13.4

FROM golang:${GO_VER}-${DEBIAN_BASE} as golang
ENV GOBIN /usr/local/bin
RUN go get -u github.com/DATA-DOG/godog/cmd/godog
RUN GO111MODULE=on go install golang.org/x/lint/golint
RUN GO111MODULE=on go install golang.org/x/tools/cmd/goimports

FROM golang:${GO_VER}-${DEBIAN_BASE}
COPY --from=golang /usr/local/bin/* /usr/local/bin/
COPY --from=golang /go/src /go/src/
RUN curl -sL https://deb.nodesource.com/setup_10.x  | bash -
RUN apt-get -y install nodejs
RUN npm install -g license-check-and-add@3.0.3
