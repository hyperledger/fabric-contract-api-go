// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/hyperledger/fabric-contract-api-go/contractapi/utils"
	"github.com/hyperledger/fabric-contract-api-go/serializer"
	"github.com/stretchr/testify/assert"
)

// ================================
// HELPERS
// ================================

type transactionHandlerStruct struct{}

func (ms *transactionHandlerStruct) BasicFunction(str string) string {
	return str
}

func (ms *transactionHandlerStruct) AdvancedFunction(str string, str2 string) string {
	return str + str2
}

func (ms *transactionHandlerStruct) GoodBeforeUnknownAfterFunction() string {
	return "CALLED GoodBeforeUnknownAfterFunction"
}

func (ms *transactionHandlerStruct) GoodBeforeUnknownAfterFunctionWithContext(ctx *TransactionContext) string {
	return "CALLED GoodBeforeUnknownAfterFunctionWithContext WITH " + ctx.Str
}

func (ms *transactionHandlerStruct) GoodAfterFunction(iface interface{}) string {
	return iface.(string)
}

func (ms *transactionHandlerStruct) GoodAfterFunctionForUndefinedInterface(iface interface{}) bool {
	_, ok := iface.(*utils.UndefinedInterface)
	return ok
}

func (ms *transactionHandlerStruct) BadFunction(param1 complex64) complex64 {
	return param1
}

type TransactionContext struct {
	Str string
}

var basicContextPtrType = reflect.TypeOf(new(TransactionContext))

// ================================
// TEST
// ================================

func TestString(t *testing.T) {
	var err error
	var str string

	str, err = TransactionHandlerTypeBefore.String()
	assert.Nil(t, err, "should output no error when before type")
	assert.Equal(t, "Before", str, "should output Before for before type")

	str, err = TransactionHandlerTypeAfter.String()
	assert.Nil(t, err, "should output no error when after type")
	assert.Equal(t, "After", str, "should output After for after type")

	str, err = TransactionHandlerTypeUnknown.String()
	assert.Nil(t, err, "should output no error when unknown type")
	assert.Equal(t, "Unknown", str, "should output Unknown for unknown type")

	str, err = TransactionHandlerType(TransactionHandlerTypeAfter + 1).String()
	assert.Error(t, err, errors.New("Invalid transaction handler type"), "should error when not one of enum")
	assert.Equal(t, "", str, "should return blank string for error")
}

func TestNewTransactionHandler(t *testing.T) {
	var th *TransactionHandler
	var err error
	var cf *ContractFunction

	ms := transactionHandlerStruct{}

	_, err = NewTransactionHandler(ms.BasicFunction, basicContextPtrType, TransactionHandlerTypeBefore)
	assert.EqualError(t, err, "Before transactions may not take any params other than the transaction context", "should error when before function takes args but not just the context")

	_, err = NewTransactionHandler(ms.BasicFunction, basicContextPtrType, TransactionHandlerTypeUnknown)
	assert.EqualError(t, err, "Unknown transactions may not take any params other than the transaction context", "should error when unknown function takes args but not just the context")

	_, err = NewTransactionHandler(ms.AdvancedFunction, basicContextPtrType, TransactionHandlerTypeAfter)
	assert.EqualError(t, err, "After transactions must take at most one non-context param", "should error when after function takes more than one non-context arg")

	_, err = NewTransactionHandler(ms.BasicFunction, basicContextPtrType, TransactionHandlerTypeAfter)
	assert.EqualError(t, err, "After transaction must take type interface{} as their only non-context param", "should error when after function takes correct number of non-context args but not interface type")

	_, expectedErr := NewContractFunctionFromFunc(ms.BadFunction, 0, basicContextPtrType)
	_, err = NewTransactionHandler(ms.BadFunction, basicContextPtrType, TransactionHandlerTypeAfter)
	assert.EqualError(t, err, fmt.Sprintf("Error creating After. %s", expectedErr.Error()), "should error when new contract function errors")

	th, err = NewTransactionHandler(ms.GoodBeforeUnknownAfterFunction, basicContextPtrType, TransactionHandlerTypeBefore)
	cf, _ = NewContractFunctionFromFunc(ms.GoodBeforeUnknownAfterFunction, 0, basicContextPtrType)
	assert.Nil(t, err, "should not error for valid tx handler (before)")
	assert.Equal(t, TransactionHandlerTypeBefore, th.handlesType, "should create a txn handler for a before txn")
	assert.Equal(t, th.params, cf.params, "should create a txn handler for a before txn that has matching contract function")
	assert.Equal(t, th.returns, cf.returns, "should create a txn handler for a before txn that has matching contract function")

	th, err = NewTransactionHandler(ms.GoodBeforeUnknownAfterFunction, basicContextPtrType, TransactionHandlerTypeUnknown)
	cf, _ = NewContractFunctionFromFunc(ms.GoodBeforeUnknownAfterFunction, 0, basicContextPtrType)
	assert.Nil(t, err, "should not error for valid tx handler (unknown)")
	assert.Equal(t, TransactionHandlerTypeUnknown, th.handlesType, "should create a txn handler for an unknown txn")
	assert.Equal(t, th.params, cf.params, "should create a txn handler for an unknown txn that has matching contract function")
	assert.Equal(t, th.returns, cf.returns, "should create a txn handler for an unknown txn that has matching contract function")

	th, err = NewTransactionHandler(ms.GoodBeforeUnknownAfterFunction, basicContextPtrType, TransactionHandlerTypeAfter)
	cf, _ = NewContractFunctionFromFunc(ms.GoodBeforeUnknownAfterFunction, 0, basicContextPtrType)
	assert.Nil(t, err, "should not error for valid tx handler (after)")
	assert.Equal(t, TransactionHandlerTypeAfter, th.handlesType, "should create a txn handler for an after txn")
	assert.Equal(t, th.params, cf.params, "should create a txn handler for an after txn that has matching contract function")
	assert.Equal(t, th.returns, cf.returns, "should create a txn handler for an after txn that has matching contract function")

	th, err = NewTransactionHandler(ms.GoodAfterFunction, basicContextPtrType, TransactionHandlerTypeAfter)
	cf, _ = NewContractFunctionFromFunc(ms.GoodAfterFunction, 0, basicContextPtrType)
	assert.Nil(t, err, "should not error for valid tx handler (afetr with param)")
	assert.Equal(t, TransactionHandlerTypeAfter, th.handlesType, "should create a txn handler for an after txn with arg")
	assert.Equal(t, th.params, cf.params, "should create a txn handler for an after txn with arg that has matching contract function")
	assert.Equal(t, th.returns, cf.returns, "should create a txn handler for an after txn with arg that has matching contract function")
}

func TestTHCall(t *testing.T) {
	var th *TransactionHandler
	var ctx = &TransactionContext{Str: "HELLO WORLD"}
	var expectedStr string
	var expectedIFace interface{}
	var expectedErr error
	var actualStr string
	var actualIFace interface{}
	var actualErr error

	serializer := new(serializer.JSONSerializer)
	ms := transactionHandlerStruct{}

	th, _ = NewTransactionHandler(ms.GoodBeforeUnknownAfterFunction, basicContextPtrType, TransactionHandlerTypeBefore)
	expectedStr, expectedIFace, expectedErr = th.handleResponse([]reflect.Value{reflect.ValueOf(ms.GoodBeforeUnknownAfterFunction())}, nil, nil, serializer)
	actualStr, actualIFace, actualErr = th.Call(reflect.ValueOf(ctx), nil, serializer)
	assert.Equal(t, expectedStr, actualStr, "should produce same string as handle response on real function")
	assert.Equal(t, expectedIFace.(string), actualIFace.(string), "should produce same interface as handle response on real function")
	assert.Equal(t, expectedErr, actualErr, "should produce same error as handle response on real function")

	th, _ = NewTransactionHandler(ms.GoodBeforeUnknownAfterFunctionWithContext, basicContextPtrType, TransactionHandlerTypeBefore)
	expectedStr, expectedIFace, expectedErr = th.handleResponse([]reflect.Value{reflect.ValueOf(ms.GoodBeforeUnknownAfterFunctionWithContext(ctx))}, nil, nil, serializer)
	actualStr, actualIFace, actualErr = th.Call(reflect.ValueOf(ctx), nil, serializer)
	assert.Equal(t, expectedStr, actualStr, "should produce same string as handle response on real function with context")
	assert.Equal(t, expectedIFace.(string), actualIFace.(string), "should produce same interface as handle response on real function with context")
	assert.Equal(t, expectedErr, actualErr, "should produce same error as handle response on real function with context")

	th, _ = NewTransactionHandler(ms.GoodAfterFunction, basicContextPtrType, TransactionHandlerTypeAfter)
	expectedStr, expectedIFace, expectedErr = th.handleResponse([]reflect.Value{reflect.ValueOf(ms.GoodAfterFunction("some str"))}, nil, nil, serializer)
	actualStr, actualIFace, actualErr = th.Call(reflect.ValueOf(ctx), "some str", serializer)
	assert.Equal(t, expectedStr, actualStr, "should produce same string as handle response on real function for after with param")
	assert.Equal(t, expectedIFace.(string), actualIFace.(string), "should produce same interface as handle response on real function for after with param")
	assert.Equal(t, expectedErr, actualErr, "should produce same error as handle response on real function for after with param")

	var ui *utils.UndefinedInterface
	th, _ = NewTransactionHandler(ms.GoodAfterFunctionForUndefinedInterface, basicContextPtrType, TransactionHandlerTypeAfter)
	expectedStr, expectedIFace, expectedErr = th.handleResponse([]reflect.Value{reflect.ValueOf(ms.GoodAfterFunctionForUndefinedInterface(ui))}, nil, nil, serializer)
	actualStr, actualIFace, actualErr = th.Call(reflect.ValueOf(ctx), nil, serializer)
	assert.Equal(t, expectedStr, actualStr, "should produce same string as handle response on real function for after with undefined interface")
	assert.Equal(t, expectedIFace.(bool), actualIFace.(bool), "should produce same interface as handle response on real function for after with undefined interface")
	assert.Equal(t, expectedErr, actualErr, "should produce same error as handle response on real function for after with undefined interface")
}
