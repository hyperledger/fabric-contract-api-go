// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import "github.com/hyperledger/fabric-contract-api-go/contractapi"

// BasicAsset holds and ID and value
type BasicAsset struct {
	ID    string `json:"id"`
	Value int    `json:"value"`
}

// Description holds details about an asset
type Description struct {
	Colour string   `json:"colour"`
	Owners []string `json:"owners"`
}

// ComplexAsset is a basic asset with more detail
type ComplexAsset struct {
	BasicAsset
	Description Description `json:"description"`
}

// AdvancedTypesContract a contract for managing non-string types in the chaincode
type AdvancedTypesContract struct {
	contractapi.Contract
}

// GetInt returns 1
func (atc *AdvancedTypesContract) GetInt() int {
	return 1
}

// CallAndResponseInt returns sent int
func (atc *AdvancedTypesContract) CallAndResponseInt(sent int) int {
	return sent
}

// GetFloat returns 1.1
func (atc *AdvancedTypesContract) GetFloat() float64 {
	return 1.1
}

// CallAndResponseFloat returns sent float
func (atc *AdvancedTypesContract) CallAndResponseFloat(sent float64) float64 {
	return sent
}

// GetBool returns true
func (atc *AdvancedTypesContract) GetBool() bool {
	return true
}

// CallAndResponseBool returns sent bool
func (atc *AdvancedTypesContract) CallAndResponseBool(sent bool) bool {
	return sent
}

// GetArray returns int array 1,2,3
func (atc *AdvancedTypesContract) GetArray() []int {
	return []int{1, 2, 3}
}

// CallAndResponseArray returns sent bool array
func (atc *AdvancedTypesContract) CallAndResponseArray(sent []bool) []bool {
	return sent
}

// GetBasicAsset returns a basic asset with id "OBJECT_1" and value 100
func (atc *AdvancedTypesContract) GetBasicAsset() BasicAsset {
	return BasicAsset{"OBJECT_1", 100}
}

// CallAndResponseBasicAsset returns sent basic asset
func (atc *AdvancedTypesContract) CallAndResponseBasicAsset(sent BasicAsset) BasicAsset {
	return sent
}

// GetComplexAsset returns a basic asset with id "OBJECT_1" and value 100
func (atc *AdvancedTypesContract) GetComplexAsset() ComplexAsset {
	return ComplexAsset{BasicAsset{"OBJECT_2", 100}, Description{"red", []string{"andy", "matthew"}}}
}

// CallAndResponseComplexAsset returns sent complex asset
func (atc *AdvancedTypesContract) CallAndResponseComplexAsset(sent ComplexAsset) ComplexAsset {
	return sent
}

func main() {
	advancedTypes := new(AdvancedTypesContract)

	cc, err := contractapi.NewChaincode(advancedTypes)

	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
