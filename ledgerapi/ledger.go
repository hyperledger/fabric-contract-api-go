// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ledgerapi

import "github.com/hyperledger/fabric-contract-api-go/contractapi"

// LedgerInterface defines functions of a ledger
type LedgerInterface interface {
	// GetCollection returns a collection with the passed name
	GetCollection(string) CollectionInterface

	// GetDefaultCollection returns a collection representing the world state
	GetDefaultCollection() CollectionInterface
}

// Ledger an implementation of LedgerInterface to provide access to collections
type Ledger struct {
	ctx contractapi.TransactionContextInterface
}

// GetCollection returns the collection identified by the passed name
func (l *Ledger) GetCollection(name string) CollectionInterface {
	collection := new(Collection)
	collection.name = name

	return collection
}

// GetDefaultCollection returns the collection identified by the WorldStateIdentifier
func (l *Ledger) GetDefaultCollection() CollectionInterface {
	return l.GetCollection(WorldStateIdentifier)
}

// GetLedger returns an implementation of the LedgerInterface setup to use the passed context
func GetLedger(ctx contractapi.TransactionContextInterface) LedgerInterface {
	ledger := new(Ledger)
	ledger.ctx = ctx

	return ledger
}
