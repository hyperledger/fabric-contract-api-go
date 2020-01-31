// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	commercialpaper "github.com/hyperledger/fabric-chaincode-integration/commercialpaper/commercial-paper"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {

	contract := new(commercialpaper.Contract)
	contract.TransactionContextHandler = new(commercialpaper.TransactionContext)
	contract.Name = "org.papernet.commercialpaper"
	contract.Info.Version = "0.0.1"

	chaincode, err := contractapi.NewChaincode(contract)

	if err != nil {
		panic(fmt.Sprintf("Error creating chaincode. %s", err.Error()))
	}

	chaincode.Info.Title = "CommercialPaperChaincode"
	chaincode.Info.Version = "0.0.1"

	err = chaincode.Start()

	if err != nil {
		panic(fmt.Sprintf("Error starting chaincode. %s", err.Error()))
	}
}
