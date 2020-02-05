// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ledgerapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLedger(t *testing.T) {
	ctx := new(MockContext)
	ctx.prop = "value"

	expectedLedger := new(Ledger)
	expectedLedger.ctx = ctx

	actualLedger := GetLedger(ctx).(*Ledger)

	assert.Equal(t, expectedLedger, actualLedger, "should return an instance of Ledger")
}

func TestGetCollection(t *testing.T) {
	ctx := new(MockContext)

	expectedCollection := new(Collection)
	expectedCollection.name = "some name"

	ledger := GetLedger(ctx)

	actualCollection := ledger.GetCollection("some name")

	assert.Equal(t, expectedCollection, actualCollection, "should return a collection using the name passed")
}

func TestGetDefaultCollection(t *testing.T) {
	ctx := new(MockContext)

	ledger := GetLedger(ctx)

	expectedCollection := ledger.GetCollection(WorldStateIdentifier)
	actualCollection := ledger.GetDefaultCollection()

	assert.Equal(t, expectedCollection, actualCollection, "should return a collection using the world state identifier")
}
