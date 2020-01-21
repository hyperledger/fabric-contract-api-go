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
# REPLACE FOLLOWING WITH DIRECT NPM INSTALL OF FABRIC-CHAINCODE-INTEGRATION
RUN mkdir -p /tmp/integration-tests
WORKDIR /tmp/integration-tests
RUN git clone --single-branch https://github.com/hyperledger/fabric-test.git
WORKDIR /tmp/integration-tests/fabric-test
RUN git checkout 8c09e15eeab7f2b9c62df47f5ef365bc185b64ba
WORKDIR /tmp/integration-tests/fabric-test/tools/chaincode-integration
RUN npm install -g fabric-chaincode-integration-0.1.0.tgz --unsafe-perm
# END OF REPLACE
RUN find /usr/lib/node_modules/fabric-chaincode-integration -type d -exec chmod 777 {} \;
RUN find /usr/lib/node_modules/fabric-chaincode-integration -type f -exec chmod 777 {} \;
RUN curl -L "https://github.com/docker/compose/releases/download/1.25.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
RUN chmod +x /usr/local/bin/docker-compose
# INSTALL DOCKER
RUN apt-get update
RUN apt-get install -y software-properties-common
RUN curl -fsSL https://download.docker.com/linux/debian/gpg | apt-key add -
RUN apt-key fingerprint 0EBFCD88
RUN add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/debian $(lsb_release -cs) stable"
RUN apt-get update
RUN apt-get -y install docker-ce
WORKDIR /go
