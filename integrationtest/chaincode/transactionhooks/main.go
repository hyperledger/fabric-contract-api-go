// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi/utils"
)

// TransactionContext a custom transaction context
type TransactionContext struct {
	contractapi.TransactionContext

	beforeKey   string
	beforeValue []byte
}

func beforeTransaction(ctx *TransactionContext) error {
	transient, _ := ctx.GetStub().GetTransient()

	if val, ok := transient["fail"]; ok && string(val) == "BEFORE" {
		return errors.New("Before transaction failed")
	}

	ctx.beforeKey = string(transient["before_key"])
	ctx.beforeValue = transient["before_value"]

	return nil
}

func afterTransaction(ctx *TransactionContext, value interface{}) error {
	transient, _ := ctx.GetStub().GetTransient()

	if val, ok := transient["fail"]; ok && string(val) == "AFTER" {
		return errors.New("After transaction failed")
	}

	if _, ok := value.(*utils.UndefinedInterface); ok {
		return nil
	}

	afterKey := string(transient["after_key"])

	return ctx.GetStub().PutState(afterKey, []byte(value.(string)))
}

// TransactionHooksContract with biz logic
type TransactionHooksContract struct {
	contractapi.Contract
}

// WriteBeforeValue writes to the world state the value stored in the context by the before function
func (thc *TransactionHooksContract) WriteBeforeValue(ctx *TransactionContext) error {
	transient, _ := ctx.GetStub().GetTransient()

	if val, ok := transient["fail"]; ok && string(val) == "NAMED" {
		return errors.New("Named transaction failed")
	}

	return ctx.GetStub().PutState(ctx.beforeKey, ctx.beforeValue)
}

// PassAfterValue returns a passed value so that it can be written to the world state by the after function
func (thc *TransactionHooksContract) PassAfterValue(ctx *TransactionContext, passToAfter string) string {
	return passToAfter
}

func main() {
	transactionhooksContract := new(TransactionHooksContract)
	transactionhooksContract.TransactionContextHandler = new(TransactionContext)
	transactionhooksContract.BeforeTransaction = beforeTransaction
	transactionhooksContract.AfterTransaction = afterTransaction

	cc, err := contractapi.NewChaincode(transactionhooksContract)

	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
