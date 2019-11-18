// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package simplecontract

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SimpleContract with biz logic
type SimpleContract struct {
	contractapi.Contract
}

// Create - Initialises a key value pair with the given ID in the world state
func (sc *SimpleContract) Create(ctx contractapi.TransactionContextInterface, key string) error {
	existing, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	if existing != nil {
		return fmt.Errorf("Cannot create key. Key with id %s already exists", key)
	}

	err = ctx.GetStub().PutState(key, []byte("Initialised"))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Update - Updates a key with given ID in the world state
func (sc *SimpleContract) Update(ctx contractapi.TransactionContextInterface, key string, value string) error {
	existing, err := ctx.GetStub().GetState(key)

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	if existing == nil {
		return fmt.Errorf("Cannot update key. Key with id %s does not exist", key)
	}

	err = ctx.GetStub().PutState(key, []byte(value))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Read - Returns value of a key with given ID from world state as string
func (sc *SimpleContract) Read(ctx contractapi.TransactionContextInterface, key string) (string, error) {
	existing, err := ctx.GetStub().GetState(key)

	if err != nil {
		return "", errors.New("Unable to interact with world state")
	}

	if existing == nil {
		return "", fmt.Errorf("Cannot read key. Key with id %s does not exist", key)
	}

	return string(existing), nil
}
