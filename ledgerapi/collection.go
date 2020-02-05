// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ledgerapi

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// WorldStateIdentifier the collection name to be used for the world state
const WorldStateIdentifier = "worldstate"

// CollectionInterface placeholder
type CollectionInterface interface {
	GetState(string) (StateInterface, error)
	GetStates(QueryInterface) (StateIteratorInterface, error)
	CreateState(string, []byte) error
	UpdateState(string, []byte) error
	DeleteState(string) error
}

// Collection placeholder
type Collection struct {
	name string
	ctx  contractapi.TransactionContextInterface
}

// GetState placeholder
func (c *Collection) GetState(key string) (StateInterface, error) {
	errMsg := fmt.Sprintf("Failed to get state %s in collection %s.", key, c.name)

	_, err := c.ctx.GetStub().GetState(key)

	if err != nil {
		return nil, fmt.Errorf("%s %s", errMsg, err.Error())
	}

	return nil, errors.New("Not yet implemented")
}

// GetStates placeholder
func (c *Collection) GetStates(QueryInterface) (StateIteratorInterface, error) {
	return nil, errors.New("Not yet implemented")
}

// CreateState adds a state with a given key to the collection. Errors if the key already exists
func (c *Collection) CreateState(key string, data []byte) error {
	errMsg := fmt.Sprintf("Failed to create new state %s in collection %s.", key, c.name)

	stub := c.ctx.GetStub()

	bytes, err := stub.GetState(key)

	if err != nil {
		return fmt.Errorf("%s %s", errMsg, err.Error())
	} else if bytes != nil {
		return fmt.Errorf("%s State already exists for key", errMsg)
	}

	err = stub.PutState(key, data)

	if err != nil {
		return fmt.Errorf("%s %s", errMsg, err.Error())
	}

	return nil
}

// UpdateState updates a the value at a the given key in the collection. Errors if that key does
// not yet exist
func (c *Collection) UpdateState(key string, data []byte) error {
	errMsg := fmt.Sprintf("Failed to update state %s in collection %s.", key, c.name)

	stub := c.ctx.GetStub()

	bytes, err := stub.GetState(key)

	if err != nil {
		return fmt.Errorf("%s %s", errMsg, err.Error())
	} else if bytes == nil {
		return fmt.Errorf("%s State does not exist for key", errMsg)
	}

	err = stub.PutState(key, data)

	if err != nil {
		return fmt.Errorf("%s %s", errMsg, err.Error())
	}

	return err
}

// DeleteState removes the value at the given key from the world state
func (c *Collection) DeleteState(key string) error {
	errMsg := fmt.Sprintf("Failed to delete state %s in collection %s.", key, c.name)

	stub := c.ctx.GetStub()

	err := stub.DelState(key)

	if err != nil {
		return fmt.Errorf("%s %s", errMsg, err.Error())
	}

	return err
}
