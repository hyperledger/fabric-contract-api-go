// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package contractapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-contract-api-go/internal"
	"github.com/hyperledger/fabric-contract-api-go/internal/utils"
	"github.com/hyperledger/fabric-contract-api-go/metadata"
	"github.com/hyperledger/fabric-contract-api-go/serializer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/assert"
)

// ================================
// HELPERS
// ================================

const standardValue = "100"
const invokeType = "INVOKE"
const initType = "INIT"
const standardTxID = "1234567890"

type simpleStruct struct {
	Prop1 string `json:"prop1"`
	prop2 string
}

func (ss *simpleStruct) GoodMethod(param1 string, param2 string) string {
	return param1 + param2
}

func (ss *simpleStruct) AnotherGoodMethod() int {
	return 1
}

type emptyContract struct {
	Contract
}

type privateContract struct {
	Contract
}

func (pc *privateContract) privateMethod() int64 {
	return 1
}

type badContract struct {
	Contract
}

func (bc *badContract) BadMethod() complex64 {
	return 1
}

type goodContract struct {
	myContract
	called []string
}

func (gc *goodContract) logBefore() {
	gc.called = append(gc.called, "Before function called")
}

func (gc *goodContract) LogNamed() string {
	gc.called = append(gc.called, "Named function called")
	return "named response"
}

func (gc *goodContract) logAfter(data interface{}) {
	gc.called = append(gc.called, fmt.Sprintf("After function called with %v", data))
}

func (gc *goodContract) logUnknown() {
	gc.called = append(gc.called, "Unknown function called")
}

func (gc *goodContract) ReturnsError() error {
	return errors.New("Some error")
}

func (gc *goodContract) ReturnsNothing() {}

func (gc *goodContract) CheckContextStub(ctx *TransactionContext) (string, error) {
	if ctx.GetStub().GetTxID() != standardTxID {
		return "", fmt.Errorf("You used a non standard txID [%s]", ctx.GetStub().GetTxID())
	}

	return "Stub as expected", nil
}

type goodContractCustomContext struct {
	Contract
}

func (sc *goodContractCustomContext) SetValInCustomContext(ctx *customContext) {
	_, params := ctx.GetStub().GetFunctionAndParameters()
	ctx.prop1 = params[0]
}

func (sc *goodContractCustomContext) GetValInCustomContext(ctx *customContext) (string, error) {
	if ctx.prop1 != standardValue {
		return "", errors.New("I wanted a standard value")
	}

	return ctx.prop1, nil
}

func (sc *goodContractCustomContext) CheckCustomContext(ctx *customContext) string {
	return ctx.ReturnString()
}

func (cc *customContext) ReturnString() string {
	return "I am custom context"
}

type ignorableFuncContract struct {
	goodContract
}

func (gifc *ignorableFuncContract) IgnoreMe() {}

func (gifc *ignorableFuncContract) GetIgnoredFunctions() []string {
	return []string{"IgnoreMe"}
}

type evaluateContract struct {
	myContract
}

func (ec *evaluateContract) GetEvaluateTransactions() []string {
	return []string{"ReturnsString"}
}

type txHandler struct{}

func (tx *txHandler) Handler() {
	// do nothing
}

func testContractChaincodeContractMatchesContract(t *testing.T, actual contractChaincodeContract, expected contractChaincodeContract) {
	t.Helper()

	assert.Equal(t, expected.info, actual.info, "should have matching info")

	if actual.beforeTransaction != nil {
		assert.Equal(t, expected.beforeTransaction.ReflectMetadata("", nil), actual.beforeTransaction.ReflectMetadata("", nil), "should have matching before transactions")
	}

	if actual.unknownTransaction != nil {
		assert.Equal(t, expected.unknownTransaction.ReflectMetadata("", nil), actual.unknownTransaction.ReflectMetadata("", nil), "should have matching before transactions")
	}

	if actual.afterTransaction != nil {
		assert.Equal(t, expected.afterTransaction.ReflectMetadata("", nil), actual.afterTransaction.ReflectMetadata("", nil), "should have matching before transactions")
	}

	assert.Equal(t, expected.transactionContextHandler, actual.transactionContextHandler, "should have matching transation contexts")

	for idx, cf := range actual.functions {
		assert.Equal(t, cf.ReflectMetadata("", nil), expected.functions[idx].ReflectMetadata("", nil), "should have matching functions")
	}
}

func callContractFunctionAndCheckError(t *testing.T, cc *ContractChaincode, arguments []string, callType string, expectedMessage string) {
	t.Helper()

	callContractFunctionAndCheckResponse(t, cc, arguments, callType, expectedMessage, "error")
}

func callContractFunctionAndCheckSuccess(t *testing.T, cc *ContractChaincode, arguments []string, callType string, expectedMessage string) {
	t.Helper()

	callContractFunctionAndCheckResponse(t, cc, arguments, callType, expectedMessage, "success")
}

func callContractFunctionAndCheckResponse(t *testing.T, cc *ContractChaincode, arguments []string, callType string, expectedMessage string, expectedType string) {
	t.Helper()

	args := [][]byte{}
	for _, str := range arguments {
		arg := []byte(str)
		args = append(args, arg)
	}

	mockStub := shimtest.NewMockStub("smartContractTest", cc)

	var response peer.Response

	if callType == initType {
		response = mockStub.MockInit(standardTxID, args)
	} else if callType == invokeType {
		response = mockStub.MockInvoke(standardTxID, args)
	} else {
		panic(fmt.Sprintf("Call type passed should be %s or %s. Value passed was %s", initType, invokeType, callType))
	}

	expectedResponse := shim.Success([]byte(expectedMessage))

	if expectedType == "error" {
		expectedResponse = shim.Error(expectedMessage)
	}

	assert.Equal(t, expectedResponse, response)
}

func testCallingContractFunctions(t *testing.T, callType string) {
	t.Helper()

	var cc *ContractChaincode

	gc := goodContract{}
	cc, _ = NewChaincode(&gc)

	// Should error when name not known
	callContractFunctionAndCheckError(t, cc, []string{"somebadname:somebadfunctionname"}, callType, "Contract not found with name somebadname")

	// should return error when function blank
	callContractFunctionAndCheckError(t, cc, []string{"goodContract:"}, callType, "Blank function name passed")

	// should return error when function not known and no unknown transaction specified
	gc.Name = "customname"
	cc, _ = NewChaincode(&gc)
	callContractFunctionAndCheckError(t, cc, []string{"customname:somebadfunctionname"}, callType, "Function somebadfunctionname not found in contract customname")

	// Should call default chaincode when name not passed
	callContractFunctionAndCheckError(t, cc, []string{"somebadfunctionname"}, callType, "Function somebadfunctionname not found in contract customname")

	gc = goodContract{}
	cc, _ = NewChaincode(&gc)

	// Should return success when function returns nothing
	callContractFunctionAndCheckSuccess(t, cc, []string{"goodContract:ReturnsNothing"}, callType, "")

	// Should return success when function starts with lower case
	callContractFunctionAndCheckSuccess(t, cc, []string{"goodContract:returnsNothing"}, callType, "")

	// should return success when function returns no error
	callContractFunctionAndCheckSuccess(t, cc, []string{"goodContract:ReturnsString"}, callType, gc.ReturnsString())

	// Should return error when function returns error
	callContractFunctionAndCheckError(t, cc, []string{"goodContract:ReturnsError"}, callType, gc.ReturnsError().Error())

	// Should return error when function unknown and set unknown function returns error
	gc.UnknownTransaction = gc.ReturnsError
	cc, _ = NewChaincode(&gc)
	callContractFunctionAndCheckError(t, cc, []string{"goodContract:somebadfunctionname"}, callType, gc.ReturnsError().Error())
	gc = goodContract{}

	// Should return success when function unknown and set unknown function returns no error
	gc.UnknownTransaction = gc.ReturnsString
	cc, _ = NewChaincode(&gc)
	callContractFunctionAndCheckSuccess(t, cc, []string{"goodContract:somebadfunctionname"}, callType, gc.ReturnsString())
	gc = goodContract{}

	// Should return error when before function returns error and not call main function
	gc.BeforeTransaction = gc.ReturnsError
	cc, _ = NewChaincode(&gc)
	callContractFunctionAndCheckError(t, cc, []string{"goodContract:ReturnsString"}, callType, gc.ReturnsError().Error())
	gc = goodContract{}

	// Should return success from passed function when before function returns no error
	gc.BeforeTransaction = gc.ReturnsString
	cc, _ = NewChaincode(&gc)
	callContractFunctionAndCheckSuccess(t, cc, []string{"goodContract:ReturnsString"}, callType, gc.ReturnsString())
	gc = goodContract{}

	// Should return error when after function returns error
	gc.AfterTransaction = gc.ReturnsError
	cc, _ = NewChaincode(&gc)
	callContractFunctionAndCheckError(t, cc, []string{"goodContract:ReturnsString"}, callType, gc.ReturnsError().Error())
	gc = goodContract{}

	// Should return success from passed function when before function returns error
	gc.AfterTransaction = gc.ReturnsString
	cc, _ = NewChaincode(&gc)
	callContractFunctionAndCheckSuccess(t, cc, []string{"goodContract:ReturnsString"}, callType, gc.ReturnsString())
	gc = goodContract{}

	// Should call before, named then after functions in order and pass name response
	gc.BeforeTransaction = gc.logBefore
	gc.AfterTransaction = gc.logAfter
	cc, _ = NewChaincode(&gc)
	callContractFunctionAndCheckSuccess(t, cc, []string{"goodContract:LogNamed"}, callType, "named response")
	assert.Equal(t, []string{"Before function called", "Named function called", "After function called with named response"}, gc.called, "Expected called field of goodContract to have logged in order before, named then after")
	gc = goodContract{}

	// Should call before, unknown then after functions in order and pass unknown response
	gc.BeforeTransaction = gc.logBefore
	gc.AfterTransaction = gc.logAfter
	gc.UnknownTransaction = gc.logUnknown
	cc, _ = NewChaincode(&gc)
	callContractFunctionAndCheckSuccess(t, cc, []string{"goodContract:somebadfunctionname"}, callType, "")
	assert.Equal(t, []string{"Before function called", "Unknown function called", "After function called with <nil>"}, gc.called, "Expected called field of goodContract to have logged in order before, named then after")
	gc = goodContract{}

	// Should pass

	// should pass the stub into transaction context as expected
	callContractFunctionAndCheckSuccess(t, cc, []string{"goodContract:CheckContextStub"}, callType, "Stub as expected")

	sc := goodContractCustomContext{}
	sc.TransactionContextHandler = new(customContext)
	cc, _ = NewChaincode(&sc)

	//should use a custom transaction context when one is set
	callContractFunctionAndCheckSuccess(t, cc, []string{"goodContractCustomContext:CheckCustomContext"}, callType, "I am custom context")

	//should use same ctx for all calls
	sc.BeforeTransaction = sc.SetValInCustomContext
	cc, _ = NewChaincode(&sc)
	callContractFunctionAndCheckSuccess(t, cc, []string{"goodContractCustomContext:GetValInCustomContext", standardValue}, callType, standardValue)

	sc.AfterTransaction = sc.GetValInCustomContext
	cc, _ = NewChaincode(&sc)
	callContractFunctionAndCheckError(t, cc, []string{"goodContractCustomContext:SetValInCustomContext", "some other value"}, callType, "I wanted a standard value")

	// should use transaction serializer
	cc, _ = NewChaincode(&gc)
	cc.TransactionSerializer = new(mockSerializer)
	callContractFunctionAndCheckSuccess(t, cc, []string{"goodContract:ReturnsString"}, callType, "GOODBYE WORLD")
}

type mockSerializer struct{}

func (ms *mockSerializer) FromString(string, reflect.Type, *metadata.ParameterMetadata, *metadata.ComponentMetadata) (reflect.Value, error) {
	return reflect.ValueOf("HELLO WORLD"), nil
}

func (ms *mockSerializer) ToString(reflect.Value, reflect.Type, *metadata.ReturnMetadata, *metadata.ComponentMetadata) (string, error) {
	return "GOODBYE WORLD", nil
}

func jsonCompare(t *testing.T, s1, s2 string) {
	t.Helper()

	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	assert.Nil(t, err, "invalid json supplied for string 1")

	err = json.Unmarshal([]byte(s2), &o2)
	assert.Nil(t, err, "invalid json supplied for string 2")

	assert.True(t, reflect.DeepEqual(o1, o2), "JSON should be equal")
}

// ================================
// TESTS
// ================================

func TestReflectMetadata(t *testing.T) {
	var reflectedMetadata metadata.ContractChaincodeMetadata

	goodMethod := new(simpleStruct).GoodMethod
	anotherGoodMethod := new(simpleStruct).AnotherGoodMethod
	ctx := reflect.TypeOf(TransactionContext{})

	info := metadata.InfoMetadata{
		Title:   "some chaincode",
		Version: "1.0.0",
	}

	cc := ContractChaincode{
		Info: info,
	}

	cf, _ := internal.NewContractFunctionFromFunc(goodMethod, internal.CallTypeEvaluate, ctx)
	cf2, _ := internal.NewContractFunctionFromFunc(anotherGoodMethod, internal.CallTypeEvaluate, ctx)

	cc.contracts = make(map[string]contractChaincodeContract)
	cc.contracts["MyContract"] = contractChaincodeContract{
		info: metadata.InfoMetadata{
			Version: "1.1.0",
			Title:   "MyContract",
		},
		functions: map[string]*internal.ContractFunction{
			"GoodMethod":        cf,
			"AnotherGoodMethod": cf2,
		},
	}

	contractMetadata := metadata.ContractMetadata{}
	contractMetadata.Name = "MyContract"
	contractMetadata.Info = new(metadata.InfoMetadata)
	contractMetadata.Info.Version = "1.1.0"
	contractMetadata.Info.Title = "MyContract"
	contractMetadata.Transactions = []metadata.TransactionMetadata{
		cf2.ReflectMetadata("AnotherGoodMethod", nil),
		cf.ReflectMetadata("GoodMethod", nil),
	} // alphabetical order
	contractMetadata.Default = false

	expectedMetadata := metadata.ContractChaincodeMetadata{}
	expectedMetadata.Info = new(metadata.InfoMetadata)
	expectedMetadata.Info.Version = "1.0.0"
	expectedMetadata.Info.Title = "some chaincode"
	expectedMetadata.Components.Schemas = make(map[string]metadata.ObjectMetadata)
	expectedMetadata.Contracts = make(map[string]metadata.ContractMetadata)
	expectedMetadata.Contracts["MyContract"] = contractMetadata

	// TESTS

	reflectedMetadata = cc.reflectMetadata()
	assert.Equal(t, expectedMetadata, reflectedMetadata, "should return contract chaincode metadata")

	expectedMetadata.Info.Version = "latest"
	cc.Info.Version = ""
	expectedMetadata.Info.Title = "undefined"
	cc.Info.Title = ""
	reflectedMetadata = cc.reflectMetadata()
	assert.Equal(t, expectedMetadata, reflectedMetadata, "should sub in value for title and version when not set")

	cc.DefaultContract = "MyContract"
	reflectedMetadata = cc.reflectMetadata()
	contractMetadata.Default = true
	expectedMetadata.Contracts["MyContract"] = contractMetadata
	assert.Equal(t, expectedMetadata, reflectedMetadata, "should return contract chaincode metadata when default")
}

func TestAugmentMetadata(t *testing.T) {
	info := metadata.InfoMetadata{
		Title:   "some chaincode",
		Version: "1.0.0",
	}

	cc := ContractChaincode{
		Info: info,
	}

	cc.augmentMetadata()

	assert.Equal(t, cc.reflectMetadata(), cc.metadata, "should return reflected metadata when none supplied as file")
}

func TestAddContract(t *testing.T) {
	var cc *ContractChaincode
	var mc *myContract
	var err error

	mc = new(myContract)
	tx := new(txHandler)

	defaultExcludes := getCiMethods()

	transactionContextPtrHandler := reflect.ValueOf(mc.GetTransactionContextHandler()).Type()

	expectedCCC := contractChaincodeContract{}
	expectedCCC.info.Version = "latest"
	expectedCCC.info.Title = "myContract"
	expectedCCC.functions = make(map[string]*internal.ContractFunction)
	expectedCCC.functions["ReturnsString"], _ = internal.NewContractFunctionFromFunc(mc.ReturnsString, internal.CallTypeSubmit, transactionContextPtrHandler)
	expectedCCC.transactionContextHandler = reflect.ValueOf(mc.GetTransactionContextHandler()).Elem().Type()

	// TESTS

	cc = new(ContractChaincode)
	cc.contracts = make(map[string]contractChaincodeContract)
	cc.contracts["customname"] = contractChaincodeContract{}
	mc = new(myContract)
	mc.Name = "customname"
	err = cc.addContract(mc, []string{})
	assert.EqualError(t, err, "Multiple contracts being merged into chaincode with name customname", "should error when contract already exists with name")

	// should error when no public functions
	cc = new(ContractChaincode)
	cc.contracts = make(map[string]contractChaincodeContract)
	ic := new(emptyContract)
	err = cc.addContract(ic, defaultExcludes)
	assert.EqualError(t, err, fmt.Sprintf("Contracts are required to have at least 1 (non-ignored) public method. Contract emptyContract has none. Method names that have been ignored: %s", utils.SliceAsCommaSentence(defaultExcludes)), "should error when contract has no public functions")

	cc = new(ContractChaincode)
	cc.contracts = make(map[string]contractChaincodeContract)
	pc := new(privateContract)
	err = cc.addContract(pc, defaultExcludes)
	assert.EqualError(t, err, fmt.Sprintf("Contracts are required to have at least 1 (non-ignored) public method. Contract privateContract has none. Method names that have been ignored: %s", utils.SliceAsCommaSentence(defaultExcludes)), "should error when contract has no public functions but private ones")

	// should add by default name
	existingCCC := contractChaincodeContract{
		info: metadata.InfoMetadata{
			Version: "some version",
		},
	}
	cc = new(ContractChaincode)
	cc.contracts = make(map[string]contractChaincodeContract)
	cc.contracts["anotherContract"] = existingCCC
	mc = new(myContract)
	err = cc.addContract(mc, defaultExcludes)
	assert.Nil(t, err, "should not error when adding contract using default name")
	assert.Equal(t, existingCCC, cc.contracts["anotherContract"], "should not affect existing contract in map")
	testContractChaincodeContractMatchesContract(t, cc.contracts["myContract"], expectedCCC)

	// should add by custom name
	cc = new(ContractChaincode)
	cc.contracts = make(map[string]contractChaincodeContract)
	mc = new(myContract)
	mc.Name = "customname"
	expectedCCC.info.Title = "customname"
	err = cc.addContract(mc, defaultExcludes)
	assert.Nil(t, err, "should not error when adding contract using custom name")
	testContractChaincodeContractMatchesContract(t, cc.contracts["customname"], expectedCCC)
	expectedCCC.info.Title = "myContract"

	// should use contracts title and version
	cc = new(ContractChaincode)
	cc.contracts = make(map[string]contractChaincodeContract)
	mc = new(myContract)
	mc.Info.Version = "1.1.0"
	mc.Info.Title = "some title"
	expectedCCC.info = metadata.InfoMetadata{
		Version: "1.1.0",
		Title:   "some title",
	}
	err = cc.addContract(mc, defaultExcludes)
	assert.Nil(t, err, "should not error when adding contract using version")
	testContractChaincodeContractMatchesContract(t, cc.contracts["myContract"], expectedCCC)
	expectedCCC.info = metadata.InfoMetadata{
		Version: "latest",
		Title:   "myContract",
	}

	// should handle evaluate functions
	cc = new(ContractChaincode)
	cc.contracts = make(map[string]contractChaincodeContract)
	oldFunc := expectedCCC.functions["ReturnsString"]
	expectedCCC.functions["ReturnsString"], _ = internal.NewContractFunctionFromFunc(mc.ReturnsString, internal.CallTypeEvaluate, transactionContextPtrHandler)
	expectedCCC.info.Title = "evaluateContract"
	ec := new(evaluateContract)
	err = cc.addContract(ec, defaultExcludes)
	assert.Nil(t, err, "should not error when adding contract using version")
	testContractChaincodeContractMatchesContract(t, cc.contracts["evaluateContract"], expectedCCC)
	expectedCCC.functions["ReturnsString"] = oldFunc
	expectedCCC.info.Title = "myContract"

	// should use before transaction
	cc = new(ContractChaincode)
	cc.contracts = make(map[string]contractChaincodeContract)
	mc = new(myContract)
	mc.BeforeTransaction = tx.Handler
	expectedCCC.beforeTransaction, _ = internal.NewTransactionHandler(tx.Handler, transactionContextPtrHandler, internal.TransactionHandlerTypeBefore)
	err = cc.addContract(mc, defaultExcludes)
	assert.Nil(t, err, "should not error when adding contract using before tx")
	testContractChaincodeContractMatchesContract(t, cc.contracts["myContract"], expectedCCC)

	// should use after transaction
	cc = new(ContractChaincode)
	cc.contracts = make(map[string]contractChaincodeContract)
	mc = new(myContract)
	mc.AfterTransaction = tx.Handler
	expectedCCC.afterTransaction, _ = internal.NewTransactionHandler(tx.Handler, transactionContextPtrHandler, internal.TransactionHandlerTypeBefore)
	err = cc.addContract(mc, defaultExcludes)
	assert.Nil(t, err, "should not error when adding contract using after tx")
	testContractChaincodeContractMatchesContract(t, cc.contracts["myContract"], expectedCCC)

	// should use unknown transaction
	cc = new(ContractChaincode)
	cc.contracts = make(map[string]contractChaincodeContract)
	mc = new(myContract)
	mc.UnknownTransaction = tx.Handler
	expectedCCC.unknownTransaction, _ = internal.NewTransactionHandler(tx.Handler, transactionContextPtrHandler, internal.TransactionHandlerTypeBefore)
	err = cc.addContract(mc, defaultExcludes)
	assert.Nil(t, err, "should not error when adding contract using unknown tx")
	testContractChaincodeContractMatchesContract(t, cc.contracts["myContract"], expectedCCC)

	// should error on bad function
	cc = new(ContractChaincode)
	cc.contracts = make(map[string]contractChaincodeContract)
	bc := new(badContract)
	err = cc.addContract(bc, defaultExcludes)
	_, expectedErr := internal.NewContractFunctionFromFunc(bc.BadMethod, internal.CallTypeSubmit, transactionContextPtrHandler)
	expectedErrStr := strings.Replace(expectedErr.Error(), "Function", "BadMethod", -1)
	assert.EqualError(t, err, expectedErrStr, "should error when contract has bad method")

	// should error on bad before transaction
	cc = new(ContractChaincode)
	cc.contracts = make(map[string]contractChaincodeContract)
	mc = new(myContract)
	mc.BeforeTransaction = bc.BadMethod
	_, expectedErr = internal.NewTransactionHandler(bc.BadMethod, transactionContextPtrHandler, internal.TransactionHandlerTypeBefore)
	err = cc.addContract(mc, defaultExcludes)
	assert.EqualError(t, err, expectedErr.Error(), "should error when before transaction is bad method")

	// should error on bad after transaction
	cc = new(ContractChaincode)
	cc.contracts = make(map[string]contractChaincodeContract)
	mc = new(myContract)
	mc.AfterTransaction = bc.BadMethod
	_, expectedErr = internal.NewTransactionHandler(bc.BadMethod, transactionContextPtrHandler, internal.TransactionHandlerTypeAfter)
	err = cc.addContract(mc, defaultExcludes)
	assert.EqualError(t, err, expectedErr.Error(), "should error when after transaction is bad method")

	// should error on bad unknown transaction
	cc = new(ContractChaincode)
	cc.contracts = make(map[string]contractChaincodeContract)
	mc = new(myContract)
	mc.UnknownTransaction = bc.BadMethod
	_, expectedErr = internal.NewTransactionHandler(bc.BadMethod, transactionContextPtrHandler, internal.TransactionHandlerTypeUnknown)
	err = cc.addContract(mc, defaultExcludes)
	assert.EqualError(t, err, expectedErr.Error(), "should error when unknown transaction is bad method")
}

func TestNewChaincode(t *testing.T) {
	var contractChaincode *ContractChaincode
	var err error
	var expectedErr error

	cc := ContractChaincode{}
	cc.contracts = make(map[string]contractChaincodeContract)

	contractChaincode, err = NewChaincode(new(badContract))
	expectedErr = cc.addContract(new(badContract), []string{})
	assert.EqualError(t, err, expectedErr.Error(), "should error when bad contract to be added")
	assert.Nil(t, contractChaincode, "should return blank contract chaincode on error")

	contractChaincode, err = NewChaincode(new(myContract), new(evaluateContract))
	assert.Nil(t, err, "should not error when passed valid contracts")
	assert.Equal(t, 3, len(contractChaincode.contracts), "should add both passed contracts and system contract")
	assert.Equal(t, reflect.TypeOf(new(serializer.JSONSerializer)), reflect.TypeOf(contractChaincode.TransactionSerializer), "should have set the transaction serializer")
	setMetadata, _, _ := contractChaincode.contracts[SystemContractName].functions["GetMetadata"].Call(reflect.ValueOf(nil), nil, nil, new(serializer.JSONSerializer))
	jsonCompare(t, "{\"info\":{\"title\":\"undefined\",\"version\":\"latest\"},\"contracts\":{\"evaluateContract\":{\"info\":{\"title\":\"evaluateContract\",\"version\":\"latest\"},\"name\":\"evaluateContract\",\"transactions\":[{\"returns\":{\"type\":\"string\"},\"tag\":[\"evaluate\"],\"name\":\"ReturnsString\"}],\"default\": false},\"myContract\":{\"info\":{\"title\":\"myContract\",\"version\":\"latest\"},\"name\":\"myContract\",\"transactions\":[{\"returns\":{\"type\":\"string\"},\"tag\":[\"submit\"],\"name\":\"ReturnsString\"}], \"default\": true},\"org.hyperledger.fabric\":{\"info\":{\"title\":\"org.hyperledger.fabric\",\"version\":\"latest\"},\"name\":\"org.hyperledger.fabric\",\"transactions\":[{\"returns\":{\"type\":\"string\"},\"tag\":[\"evaluate\"],\"name\":\"GetMetadata\"}], \"default\": false}},\"components\":{}}", setMetadata)

	contractChaincode, err = NewChaincode(new(ignorableFuncContract))
	_, ok := contractChaincode.contracts["ignorableFuncContract"].functions["IgnoreMe"]
	assert.Nil(t, err, "should not return error for valid contract with ignores")
	assert.False(t, ok, "should not include ignored function")
}

func TestStart(t *testing.T) {
	mc := new(myContract)

	cc, _ := NewChaincode(mc)

	assert.EqualError(t, cc.Start(), shim.Start(cc).Error(), "should call shim.Start()")
}

func TestInit(t *testing.T) {
	cc, _ := NewChaincode(new(myContract))
	mockStub := shimtest.NewMockStub("blank fcn", cc)
	assert.Equal(t, shim.Success([]byte("Default initiator successful.")), cc.Init(mockStub), "should just return success on init with no function passed")

	testCallingContractFunctions(t, initType)
}

func TestInvoke(t *testing.T) {
	testCallingContractFunctions(t, invokeType)
}
