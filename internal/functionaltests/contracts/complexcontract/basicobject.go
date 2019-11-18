// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package complexcontract

// BasicOwner details about an owner
type BasicOwner struct {
	Name    string `json:"name"`
	Contact string `json:"contact"`
}

// BasicObject a basic object
type BasicObject struct {
	ID        string     `json:"id"`
	Owner     BasicOwner `json:"owner"`
	Value     uint       `json:"value"`
	Condition int        `json:"condition"`
	Colours   []string   `json:"colours"`
}

// SetConditionNew set the condition of the object to mark as new
func (ba *BasicObject) SetConditionNew() {
	ba.Condition = 0
}

// SetConditionUsed set the condition of the object to mark as used
func (ba *BasicObject) SetConditionUsed() {
	ba.Condition = 1
}
