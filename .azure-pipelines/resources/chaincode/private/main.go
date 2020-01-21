// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// PrivateContract with biz logic
type PrivateContract struct {
	contractapi.Contract
}

// PutPrivateState - Adds a key value pair to the private collection Org1AndOrg2
func (sc *PrivateContract) PutPrivateState(ctx contractapi.TransactionContextInterface, key string, value string) error {
	return ctx.GetStub().PutPrivateData("Org1AndOrg2", key, []byte(value))
}

// GetPrivateState - Gets the value for a key from the private collection Org1AndOrg2
func (sc *PrivateContract) GetPrivateState(ctx contractapi.TransactionContextInterface, key string) (string, error) {
	bytes, err := ctx.GetStub().GetPrivateData("Org1AndOrg2", key)

	if err != nil {
		return "", errors.New("Could not read private collection Org1AndOrg2")
	}

	if bytes == nil {
		return "", errors.New("No value found for " + key)
	}

	return string(bytes), nil
}

// DeletePrivateState - Deletes a key from the private collection Org1AndOrg2
func (sc *PrivateContract) DeletePrivateState(ctx contractapi.TransactionContextInterface, key string) error {
	return ctx.GetStub().DelPrivateData("Org1AndOrg2", key)
}

func main() {
	privateContract := new(PrivateContract)

	cc, err := contractapi.NewChaincode(privateContract)

	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
