// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SimpleContract with biz logic
type SimpleContract struct {
	contractapi.Contract
}

// HelloWorld - returns a string
func (sc *SimpleContract) HelloWorld(ctx contractapi.TransactionContextInterface) string {
	return "Hello World"
}

// CallAndResponse - Returns the string you send
func (sc *SimpleContract) CallAndResponse(ctx contractapi.TransactionContextInterface, value string) string {
	return value
}

// PutState - Adds a key value pair to the world state
func (sc *SimpleContract) PutState(ctx contractapi.TransactionContextInterface, key string, value string) error {
	return ctx.GetStub().PutState(key, []byte(value))
}

// GetState - Gets the value for a key from the world state
func (sc *SimpleContract) GetState(ctx contractapi.TransactionContextInterface, key string) (string, error) {
	bytes, err := ctx.GetStub().GetState(key)

	if err != nil {
		return "", nil
	}

	return string(bytes), nil
}

// DeleteState - Deletes a key from the world state
func (sc *SimpleContract) DeleteState(ctx contractapi.TransactionContextInterface, key string) error {
	return ctx.GetStub().DelState(key)
}

func main() {
	simpleContract := new(SimpleContract)

	cc, err := contractapi.NewChaincode(simpleContract)

	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
