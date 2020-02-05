// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ledgerapi

import (
	"errors"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/stretchr/testify/mock"
)

const getStateError = "collection get error"
const putStateError = "collection put error"
const delStateError = "collection del error"

type MockStub struct {
	shim.ChaincodeStubInterface
	mock.Mock
}

func (ms *MockStub) GetState(key string) ([]byte, error) {
	args := ms.Called(key)

	return args.Get(0).([]byte), args.Error(1)
}

func (ms *MockStub) PutState(key string, value []byte) error {
	args := ms.Called(key, value)

	return args.Error(0)
}

func (ms *MockStub) DelState(key string) error {
	args := ms.Called(key)

	return args.Error(0)
}

type MockContext struct {
	mock.Mock
	prop string
}

func (mc *MockContext) GetStub() shim.ChaincodeStubInterface {
	args := mc.Called()

	return args.Get(0).(*MockStub)
}

func (mc *MockContext) GetClientIdentity() cid.ClientIdentity {
	return nil
}

func configureStub() (*MockContext, *MockStub) {
	var nilBytes []byte

	ms := new(MockStub)
	ms.On("GetState", "readbad").Return(nilBytes, errors.New(getStateError))
	ms.On("GetState", "missingkey").Return(nilBytes, nil)
	ms.On("GetState", "existingkey").Return([]byte("some value"), nil)
	ms.On("GetState", "putbadmissing").Return(nilBytes, nil)
	ms.On("GetState", "putbadexisting").Return([]byte("some value"), nil)

	ms.On("PutState", "putbadmissing", mock.AnythingOfType("[]uint8")).Return(errors.New(putStateError))
	ms.On("PutState", "putbadexisting", mock.AnythingOfType("[]uint8")).Return(errors.New(putStateError))
	ms.On("PutState", "missingkey", mock.AnythingOfType("[]uint8")).Return(nil)
	ms.On("PutState", "existingkey", mock.AnythingOfType("[]uint8")).Return(nil)

	ms.On("DelState", "delbad").Return(errors.New(delStateError))
	ms.On("DelState", "existingkey").Return(nil)

	mc := new(MockContext)
	mc.On("GetStub").Return(ms)

	return mc, ms
}
