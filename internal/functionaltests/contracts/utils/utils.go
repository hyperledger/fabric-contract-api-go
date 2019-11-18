// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// CustomTransactionContextInterface adds extra methods to basic context
// interface to give access to callData
type CustomTransactionContextInterface interface {
	contractapi.TransactionContextInterface

	SetCallData([]byte)
	GetCallData() []byte
}

// CustomTransactionContext adds extra field to contractapi.TransactionContext
// so that data can be between calls
type CustomTransactionContext struct {
	contractapi.TransactionContext
	callData []byte
}

// SetCallData sets the call data property
func (ctx *CustomTransactionContext) SetCallData(bytes []byte) {
	ctx.callData = bytes
}

// GetCallData gets the call data property
func (ctx *CustomTransactionContext) GetCallData() []byte {
	return ctx.callData
}

// GetWorldState takes a key and sets what is found in the world state for that
// key in the transaction context
func GetWorldState(ctx CustomTransactionContextInterface) error {
	_, params := ctx.GetStub().GetFunctionAndParameters()

	if len(params) < 1 {
		return errors.New("Missing key for world state")
	}

	existing, err := ctx.GetStub().GetState(params[0])

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	ctx.SetCallData(existing)

	return nil
}

// UnknownTransactionHandler logs details of a bad transaction request
// and returns a shim error
func UnknownTransactionHandler(ctx CustomTransactionContextInterface) error {
	fcn, args := ctx.GetStub().GetFunctionAndParameters()
	return fmt.Errorf("Invalid function %s passed with args [%s]", fcn, strings.Join(args, ", "))
}
