// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ledgerapi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ================================
// HELPERS
// ================================

func configureCollection() *Collection {
	ctx, _ := configureStub()

	collection := new(Collection)
	collection.ctx = ctx
	collection.name = "mycollection"

	return collection
}

// ================================
// TESTS
// ================================

func TestGetState(t *testing.T) {
	collection := configureCollection()

	si, err := collection.GetState("readbad")

	assert.EqualError(t, err, fmt.Sprintf("Failed to get state readbad in collection mycollection. %s", getStateError), "should error when stub GetState errors")
	assert.Nil(t, si, "should not return state interface when stub GetState fails to return a value")
}

func TestGetStates(t *testing.T) {
	// not yet implemented
}

func TestCreateState(t *testing.T) {
	collection := configureCollection()
	stub := (collection.ctx.GetStub().(*MockStub))

	var err error

	err = collection.CreateState("readbad", []byte("value"))
	assert.EqualError(t, err, fmt.Sprintf("Failed to create new state readbad in collection mycollection. %s", getStateError), "should error when stub GetState errors")

	err = collection.CreateState("existingkey", []byte("value"))
	assert.EqualError(t, err, "Failed to create new state existingkey in collection mycollection. State already exists for key", "should error when key exists")
	stub.AssertNotCalled(t, "PutState", "existingkey", []byte("value"))

	err = collection.CreateState("putbadmissing", []byte("value"))
	assert.EqualError(t, err, fmt.Sprintf("Failed to create new state putbadmissing in collection mycollection. %s", putStateError), "should error when stub PutState errors")

	err = collection.CreateState("missingkey", []byte("value"))
	assert.Nil(t, err, "should not error when key is new")
	stub.AssertCalled(t, "PutState", "missingkey", []byte("value"))
}

func TestUpdateState(t *testing.T) {
	collection := configureCollection()
	stub := (collection.ctx.GetStub().(*MockStub))

	var err error

	err = collection.UpdateState("readbad", []byte("value"))
	assert.EqualError(t, err, fmt.Sprintf("Failed to update state readbad in collection mycollection. %s", getStateError), "should error when stub GetState errors")

	err = collection.UpdateState("missingkey", []byte("value"))
	assert.EqualError(t, err, "Failed to update state missingkey in collection mycollection. State does not exist for key", "should error when key exists")
	stub.AssertNotCalled(t, "PutState", "missingkey", []byte("value"))

	err = collection.UpdateState("putbadexisting", []byte("value"))
	assert.EqualError(t, err, fmt.Sprintf("Failed to update state putbadexisting in collection mycollection. %s", putStateError), "should error when stub PutState errors")

	err = collection.UpdateState("existingkey", []byte("value"))
	assert.Nil(t, err, "should not error when key exists")
	stub.AssertCalled(t, "PutState", "existingkey", []byte("value"))
}

func TestDeleteState(t *testing.T) {
	collection := configureCollection()
	stub := (collection.ctx.GetStub().(*MockStub))

	var err error

	err = collection.DeleteState("delbad")
	assert.EqualError(t, err, fmt.Sprintf("Failed to delete state delbad in collection mycollection. %s", delStateError), "should error when stub DelState errors")

	err = collection.DeleteState("existingkey")
	assert.Nil(t, err, "should not error when can delete key")
	stub.AssertCalled(t, "DelState", "existingkey")
}
