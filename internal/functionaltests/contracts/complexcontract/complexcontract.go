// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package complexcontract

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-contract-api-go/internal/functionaltests/contracts/utils"
)

// ComplexContract contract for handling the business logic of a basic object
type ComplexContract struct {
	contractapi.Contract
}

// NewObject adds a new basic object to the world state using id as key
func (c *ComplexContract) NewObject(ctx utils.CustomTransactionContextInterface, id string, owner BasicOwner, value uint, colours []string) error {
	existing := ctx.GetCallData()

	if existing != nil {
		return fmt.Errorf("Cannot create new object in world state as key %s already exists", id)
	}

	ba := BasicObject{}
	ba.ID = id
	ba.Owner = owner
	ba.Value = value
	ba.Colours = colours
	ba.SetConditionNew()

	baBytes, _ := json.Marshal(ba)

	err := ctx.GetStub().PutState(id, []byte(baBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// UpdateOwner changes the ownership of a basic object and mark it as used
func (c *ComplexContract) UpdateOwner(ctx utils.CustomTransactionContextInterface, id string, newOwner BasicOwner) error {
	existing := ctx.GetCallData()

	if existing == nil {
		return fmt.Errorf("Cannot update object in world state as key %s does not exist", id)
	}

	ba := BasicObject{}

	err := json.Unmarshal(existing, &ba)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type BasicObject", id)
	}

	ba.Owner = newOwner
	ba.SetConditionUsed()

	baBytes, _ := json.Marshal(ba)

	err = ctx.GetStub().PutState(id, []byte(baBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// UpdateValue changes the value of a basic object to add the value passed
func (c *ComplexContract) UpdateValue(ctx utils.CustomTransactionContextInterface, id string, valueAdd int) error {
	existing := ctx.GetCallData()

	if existing == nil {
		return fmt.Errorf("Cannot update object in world state as key %s does not exist", id)
	}

	ba := BasicObject{}

	err := json.Unmarshal(existing, &ba)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type BasicObject", id)
	}

	newValue := int(ba.Value) + valueAdd

	if newValue < 0 {
		newValue = 0
	}

	ba.Value = uint(newValue)

	baBytes, _ := json.Marshal(ba)

	err = ctx.GetStub().PutState(id, []byte(baBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// GetObject returns the object with id given from the world state
func (c *ComplexContract) GetObject(ctx utils.CustomTransactionContextInterface, id string) (*BasicObject, error) {
	existing := ctx.GetCallData()

	if existing == nil {
		return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", id)
	}

	ba := new(BasicObject)

	err := json.Unmarshal(existing, ba)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type BasicObject", id)
	}

	return ba, nil
}

// GetValue returns the value from the object with id given from the world state
func (c *ComplexContract) GetValue(ctx utils.CustomTransactionContextInterface, id string) (uint, error) {
	existing := ctx.GetCallData()

	if existing == nil {
		return 0, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", id)
	}

	ba := new(BasicObject)

	err := json.Unmarshal(existing, ba)

	if err != nil {
		return 0, fmt.Errorf("Data retrieved from world state for key %s was not of type BasicObject", id)
	}

	return ba.Value, nil
}
