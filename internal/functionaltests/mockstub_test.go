// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package functionaltests

import (
	"errors"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	minUnicodeRuneValue = 0 //U+0000
)

// MockStub is an implementation of ChaincodeStubInterface for unit testing chaincode.
// Use this instead of ChaincodeStub in your chaincode's unit test calls to Init or Invoke.
type MockStub struct {
	// arguments the stub was called with
	args [][]byte

	// A pointer back to the chaincode that will invoke this, set by constructor.
	// If a peer calls this stub, the chaincode will be invoked from here.
	cc shim.Chaincode

	// State keeps name value pairs
	State map[string][]byte

	// stores a transaction uuid while being Invoked / Deployed
	// TODO if a chaincode uses recursion this may need to be a stack of TxIDs or possibly a reference counting map
	TxID string

	TxTimestamp *timestamppb.Timestamp

	// mocked signedProposal
	signedProposal *peer.SignedProposal

	// stores a channel ID of the proposal
	ChannelID string
}

// GetTxID ...
func (stub *MockStub) GetTxID() string {
	return stub.TxID
}

// GetChannelID ...
func (stub *MockStub) GetChannelID() string {
	return stub.ChannelID
}

// GetArgs ...
func (stub *MockStub) GetArgs() [][]byte {
	return stub.args
}

// GetStringArgs ...
func (stub *MockStub) GetStringArgs() []string {
	args := stub.GetArgs()
	strargs := make([]string, 0, len(args))
	for _, barg := range args {
		strargs = append(strargs, string(barg))
	}
	return strargs
}

// GetFunctionAndParameters ...
func (stub *MockStub) GetFunctionAndParameters() (function string, params []string) {
	allargs := stub.GetStringArgs()
	function = ""
	params = []string{}
	if len(allargs) >= 1 {
		function = allargs[0]
		params = allargs[1:]
	}
	return
}

// MockTransactionStart Used to indicate to a chaincode that it is part of a transaction.
// This is important when chaincodes invoke each other.
// MockStub doesn't support concurrent transactions at present.
func (stub *MockStub) MockTransactionStart(txid string) {
	stub.TxID = txid
	stub.setSignedProposal(&peer.SignedProposal{})
	stub.TxTimestamp = timestamppb.Now()
}

// MockTransactionEnd End a mocked transaction, clearing the UUID.
func (stub *MockStub) MockTransactionEnd(uuid string) {
	stub.signedProposal = nil
	stub.TxID = ""
}

// MockInit Initialise this chaincode, also starts and ends a transaction.
func (stub *MockStub) MockInit(txID string, args [][]byte) *peer.Response {
	stub.args = args
	stub.MockTransactionStart(txID)
	res := stub.cc.Init(stub)
	stub.MockTransactionEnd(txID)
	return res
}

// MockInvoke Invoke this chaincode, also starts and ends a transaction.
func (stub *MockStub) MockInvoke(txID string, args [][]byte) *peer.Response {
	stub.args = args
	stub.MockTransactionStart(txID)
	res := stub.cc.Invoke(stub)
	stub.MockTransactionEnd(txID)
	return res
}

// GetDecorations ...
func (stub *MockStub) GetDecorations() map[string][]byte {
	return nil
}

// MockInvokeWithSignedProposal Invoke this chaincode, also starts and ends a transaction.
func (stub *MockStub) MockInvokeWithSignedProposal(txID string, args [][]byte, sp *peer.SignedProposal) *peer.Response {
	stub.args = args
	stub.MockTransactionStart(txID)
	stub.signedProposal = sp
	res := stub.cc.Invoke(stub)
	stub.MockTransactionEnd(txID)
	return res
}

// GetPrivateData ...
func (stub *MockStub) GetPrivateData(collection string, key string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// GetPrivateDataHash ...
func (stub *MockStub) GetPrivateDataHash(collection, key string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// PutPrivateData ...
func (stub *MockStub) PutPrivateData(collection string, key string, value []byte) error {
	return errors.New("not implemented")
}

// DelPrivateData ...
func (stub *MockStub) DelPrivateData(collection string, key string) error {
	return errors.New("not implemented")
}

// PurgePrivateData ...
func (stub *MockStub) PurgePrivateData(collection string, key string) error {
	return errors.New("not implemented")
}

// GetPrivateDataByRange ...
func (stub *MockStub) GetPrivateDataByRange(collection, startKey, endKey string) (shim.StateQueryIteratorInterface, error) {
	return nil, errors.New("not implemented")
}

// GetPrivateDataByPartialCompositeKey ...
func (stub *MockStub) GetPrivateDataByPartialCompositeKey(collection, objectType string, attributes []string) (shim.StateQueryIteratorInterface, error) {
	return nil, errors.New("not implemented")
}

// GetPrivateDataQueryResult ...
func (stub *MockStub) GetPrivateDataQueryResult(collection, query string) (shim.StateQueryIteratorInterface, error) {
	return nil, errors.New("not implemented")
}

// GetState retrieves the value for a given key from the ledger
func (stub *MockStub) GetState(key string) ([]byte, error) {
	value := stub.State[key]
	return value, nil
}

// PutState writes the specified `value` and `key` into the ledger.
func (stub *MockStub) PutState(key string, value []byte) error {
	if stub.TxID == "" {
		err := errors.New("cannot PutState without a transactions - call stub.MockTransactionStart()?")
		return err
	}

	// If the value is nil or empty, delete the key
	if len(value) == 0 {
		return stub.DelState(key)
	}

	stub.State[key] = value

	return nil
}

// DelState removes the specified `key` and its value from the ledger.
func (stub *MockStub) DelState(key string) error {
	delete(stub.State, key)
	return nil
}

// GetStateByRange ...
func (stub *MockStub) GetStateByRange(startKey, endKey string) (shim.StateQueryIteratorInterface, error) {
	return nil, errors.New("not implemented")
}

// GetQueryResult function can be invoked by a chaincode to perform a
// rich query against state database.  Only supported by state database implementations
// that support rich query.  The query string is in the syntax of the underlying
// state database. An iterator is returned which can be used to iterate (next) over
// the query result set
func (stub *MockStub) GetQueryResult(query string) (shim.StateQueryIteratorInterface, error) {
	// Not implemented since the mock engine does not have a query engine.
	// However, a very simple query engine that supports string matching
	// could be implemented to test that the framework supports queries
	return nil, errors.New("not implemented")
}

// GetHistoryForKey function can be invoked by a chaincode to return a history of
// key values across time. GetHistoryForKey is intended to be used for read-only queries.
func (stub *MockStub) GetHistoryForKey(key string) (shim.HistoryQueryIteratorInterface, error) {
	return nil, errors.New("not implemented")
}

// GetStateByPartialCompositeKey function can be invoked by a chaincode to query the
// state based on a given partial composite key. This function returns an
// iterator which can be used to iterate over all composite keys whose prefix
// matches the given partial composite key. This function should be used only for
// a partial composite key. For a full composite key, an iter with empty response
// would be returned.
func (stub *MockStub) GetStateByPartialCompositeKey(objectType string, attributes []string) (shim.StateQueryIteratorInterface, error) {
	return nil, errors.New("not implemented")
}

// CreateCompositeKey combines the list of attributes
// to form a composite key.
func (stub *MockStub) CreateCompositeKey(objectType string, attributes []string) (string, error) {
	return shim.CreateCompositeKey(objectType, attributes)
}

// SplitCompositeKey splits the composite key into attributes
// on which the composite key was formed.
func (stub *MockStub) SplitCompositeKey(compositeKey string) (string, []string, error) {
	return splitCompositeKey(compositeKey)
}

func splitCompositeKey(compositeKey string) (string, []string, error) {
	componentIndex := 1
	components := []string{}
	for i := 1; i < len(compositeKey); i++ {
		if compositeKey[i] == minUnicodeRuneValue {
			components = append(components, compositeKey[componentIndex:i])
			componentIndex = i + 1
		}
	}
	return components[0], components[1:], nil
}

// GetStateByRangeWithPagination ...
func (stub *MockStub) GetStateByRangeWithPagination(startKey, endKey string, pageSize int32,
	bookmark string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	return nil, nil, errors.New("not implemented")
}

// GetStateByPartialCompositeKeyWithPagination ...
func (stub *MockStub) GetStateByPartialCompositeKeyWithPagination(objectType string, keys []string,
	pageSize int32, bookmark string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	return nil, nil, errors.New("not implemented")
}

// GetQueryResultWithPagination ...
func (stub *MockStub) GetQueryResultWithPagination(query string, pageSize int32,
	bookmark string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	return nil, nil, nil
}

// InvokeChaincode locally calls the specified chaincode `Invoke`.
// E.g. stub1.InvokeChaincode("othercc", funcArgs, channel)
// Before calling this make sure to create another MockStub stub2, call shim.NewMockStub("othercc", Chaincode)
// and register it with stub1 by calling stub1.MockPeerChaincode("othercc", stub2, channel)
func (stub *MockStub) InvokeChaincode(chaincodeName string, args [][]byte, channel string) *peer.Response {
	return nil
}

// GetCreator ...
func (stub *MockStub) GetCreator() ([]byte, error) {
	return nil, errors.New("not implemented")
}

// SetTransient set TransientMap to mockStub
func (stub *MockStub) SetTransient(tMap map[string][]byte) error {
	return errors.New("not implemented")
}

// GetTransient ...
func (stub *MockStub) GetTransient() (map[string][]byte, error) {
	return nil, errors.New("not implemented")
}

// GetBinding Not implemented ...
func (stub *MockStub) GetBinding() ([]byte, error) {
	return nil, errors.New("not implemented")
}

// GetSignedProposal Not implemented ...
func (stub *MockStub) GetSignedProposal() (*peer.SignedProposal, error) {
	return stub.signedProposal, nil
}

func (stub *MockStub) setSignedProposal(sp *peer.SignedProposal) {
	stub.signedProposal = sp
}

// GetArgsSlice Not implemented ...
func (stub *MockStub) GetArgsSlice() ([]byte, error) {
	return nil, errors.New("not implemented")
}

// GetTxTimestamp ...
func (stub *MockStub) GetTxTimestamp() (*timestamppb.Timestamp, error) {
	if stub.TxTimestamp == nil {
		return nil, errors.New("timestamp not set")
	}
	return stub.TxTimestamp, nil
}

// SetEvent ...
func (stub *MockStub) SetEvent(name string, payload []byte) error {
	return errors.New("not implemented")
}

// SetStateValidationParameter ...
func (stub *MockStub) SetStateValidationParameter(key string, ep []byte) error {
	return errors.New("not implemented")
}

// GetStateValidationParameter ...
func (stub *MockStub) GetStateValidationParameter(key string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// SetPrivateDataValidationParameter ...
func (stub *MockStub) SetPrivateDataValidationParameter(collection, key string, ep []byte) error {
	return errors.New("not implemented")
}

// GetPrivateDataValidationParameter ...
func (stub *MockStub) GetPrivateDataValidationParameter(collection, key string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (stub *MockStub) FinishWriteBatch() error {
	return errors.New("not implemented")
}

func (stub *MockStub) GetAllStatesCompositeKeyWithPagination(pageSize int32,
	bookmark string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	return nil, nil, errors.New("not implemented")
}

func (stub *MockStub) GetMultiplePrivateData(collection string, keys ...string) ([][]byte, error) {
	return nil, errors.New("not implemented")
}

func (stub *MockStub) GetMultipleStates(keys ...string) ([][]byte, error) {
	return nil, errors.New("not implemented")
}

func (stub *MockStub) StartWriteBatch() {}

// NewMockStub Constructor to initialise the internal State map
func NewMockStub(cc shim.Chaincode) *MockStub {
	return &MockStub{
		cc:    cc,
		State: make(map[string][]byte),
	}
}
