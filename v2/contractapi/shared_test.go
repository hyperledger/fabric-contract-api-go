// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package contractapi

type myContract struct {
	Contract
	//lint:ignore U1000 unused
	called []string
}

func (mc *myContract) ReturnsString() string {
	return "Some string"
}

type customContext struct {
	TransactionContext
	prop1 string
}
