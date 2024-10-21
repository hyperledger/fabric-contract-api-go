// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi/utils"
	"github.com/hyperledger/fabric-contract-api-go/v2/serializer"
)

// TransactionHandlerType enum for type of transaction handled
type TransactionHandlerType int

const (
	// TransactionHandlerTypeBefore before transaction type
	TransactionHandlerTypeBefore TransactionHandlerType = iota + 1
	// TransactionHandlerTypeUnknown before transaction type
	TransactionHandlerTypeUnknown
	// TransactionHandlerTypeAfter before transaction type
	TransactionHandlerTypeAfter
)

var ErrInvalidTransactionHandlerType = errors.New("invalid transaction handler type")

func (tht TransactionHandlerType) String() string {
	switch tht {
	case TransactionHandlerTypeBefore:
		return "Before"
	case TransactionHandlerTypeAfter:
		return "After"
	case TransactionHandlerTypeUnknown:
		return "Unknown"
	default:
		return "Invalid"
	}
}

// TransactionHandler extension of contract function that manages function which handles calls
// to before, after and unknown transaction functions
type TransactionHandler struct {
	ContractFunction
	handlesType TransactionHandlerType
}

// Call calls transaction function using string args and handles formatting the response into useful types
func (th TransactionHandler) Call(ctx reflect.Value, data interface{}, serializer serializer.TransactionSerializer) (string, interface{}, error) {
	values := make([]reflect.Value, 0, 2)

	if th.params.context != nil {
		values = append(values, ctx)
	}

	if th.handlesType == TransactionHandlerTypeAfter && len(th.params.fields) == 1 {
		if data == nil {
			values = append(values, reflect.Zero(reflect.TypeOf((*utils.UndefinedInterface)(nil)).Elem()))
		} else {
			values = append(values, reflect.ValueOf(data))
		}
	}

	response := th.function.Call(values)

	return th.handleResponse(response, nil, nil, serializer)
}

// NewTransactionHandler create a new transaction handler from a given function
func NewTransactionHandler(fn interface{}, contextHandlerType reflect.Type, handlesType TransactionHandlerType) (*TransactionHandler, error) {
	cf, err := NewContractFunctionFromFunc(fn, 0, contextHandlerType)
	if err != nil {
		return nil, fmt.Errorf("error creating %s: %w", handlesType, err)
	}

	if err := validateTransactionHandler(cf, handlesType); err != nil {
		return nil, err
	}

	return &TransactionHandler{
		ContractFunction: *cf,
		handlesType:      handlesType,
	}, nil
}

func validateTransactionHandler(cf *ContractFunction, handlesType TransactionHandlerType) error {
	if handlesType != TransactionHandlerTypeAfter && len(cf.params.fields) > 0 {
		return fmt.Errorf("%s transactions may not take any params other than the transaction context", handlesType)
	}

	if handlesType == TransactionHandlerTypeAfter {
		if len(cf.params.fields) > 1 {
			return errors.New("after transactions must take at most one non-context param")
		}
		if len(cf.params.fields) == 1 && cf.params.fields[0].Kind() != reflect.Interface {
			return errors.New("after transaction must take type interface{} as their only non-context param")
		}
	}

	return nil
}
