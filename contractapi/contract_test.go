// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package contractapi

import (
	"testing"

	"github.com/hyperledger/fabric-contract-api-go/metadata"
	"github.com/stretchr/testify/assert"
)

// ================================
// Tests
// ================================

func TestGetUnknownTransaction(t *testing.T) {
	var mc myContract
	var unknownFn interface{}

	mc = myContract{}
	unknownFn = mc.GetUnknownTransaction()
	assert.Nil(t, unknownFn, "should not return contractFunction when unknown transaction not set")

	mc = myContract{}
	mc.UnknownTransaction = mc.ReturnsString
	unknownFn = mc.GetUnknownTransaction()
	assert.Equal(t, mc.ReturnsString(), unknownFn.(func() string)(), "function returned should be same value as set for unknown transaction")
}

func TestGetBeforeTransaction(t *testing.T) {
	var mc myContract
	var beforeFn interface{}

	mc = myContract{}
	beforeFn = mc.GetBeforeTransaction()
	assert.Nil(t, beforeFn, "should not return contractFunction when before transaction not set")

	mc = myContract{}
	mc.BeforeTransaction = mc.ReturnsString
	beforeFn = mc.GetBeforeTransaction()
	assert.Equal(t, mc.ReturnsString(), beforeFn.(func() string)(), "function returned should be same value as set for before transaction")
}

func TestGetAfterTransaction(t *testing.T) {
	var mc myContract
	var afterFn interface{}

	mc = myContract{}
	afterFn = mc.GetAfterTransaction()
	assert.Nil(t, afterFn, "should not return contractFunction when after transaction not set")

	mc = myContract{}
	mc.AfterTransaction = mc.ReturnsString
	afterFn = mc.GetAfterTransaction()
	assert.Equal(t, mc.ReturnsString(), afterFn.(func() string)(), "function returned should be same value as set for after transaction")
}

func TestGetInfo(t *testing.T) {
	c := Contract{}
	c.Info = metadata.InfoMetadata{}
	c.Info.Version = "some version"

	assert.Equal(t, c.Info, c.GetInfo(), "should set the version")
}

func TestGetName(t *testing.T) {
	mc := myContract{}

	assert.Equal(t, "", mc.GetName(), "should have returned blank ns when not set")

	mc.Name = "myname"
	assert.Equal(t, "myname", mc.GetName(), "should have returned custom ns when set")
}

func TestGetTransactionContextHandler(t *testing.T) {
	mc := myContract{}

	assert.Equal(t, new(TransactionContext), mc.GetTransactionContextHandler(), "should return default transaction context type when unset")

	mc.TransactionContextHandler = new(customContext)
	assert.Equal(t, new(customContext), mc.GetTransactionContextHandler(), "should return custom context when set")
}
