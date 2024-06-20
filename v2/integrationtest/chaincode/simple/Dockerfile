# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

FROM hyperledger/fabric-contract-api-go-integrationtest

WORKDIR $GOPATH/src/github.com/hyperledger/fabric-contract-api-go/integrationtest/chaincode/simple
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 9999
CMD ["/go/bin/simple"]
