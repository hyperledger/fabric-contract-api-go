// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-openapi/spec"
	metadata "github.com/hyperledger/fabric-contract-api-go/metadata"
	"github.com/hyperledger/fabric-contract-api-go/serializer"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

// ================================
// HELPERS
// ================================

type TransactionInterface interface {
	SomeFunction(string) string
}

func (tc *TransactionContext) SomeFunction(param0 string) string {
	return ""
}

type BadTransactionInterface interface {
	SomeOtherFunction() string
}

type simpleStruct struct {
	Prop1 string `json:"prop1"`
	prop2 string
}

func (ss *simpleStruct) GoodMethod(param1 string, param2 string) string {
	return param1 + param2
}

func (ss *simpleStruct) GoodTransactionMethod(ctx *TransactionContext, param1 string, param2 string) string {
	return param1 + param2
}

func (ss *simpleStruct) GoodTransactionInterfaceMethod(ctx TransactionInterface, param1 string, param2 string) string {
	return param1 + param2
}

func (ss *simpleStruct) GoodReturnMethod(param1 string) (string, error) {
	return param1, nil
}

func (ss *simpleStruct) GoodErrorMethod() error {
	return nil
}

func (ss *simpleStruct) GoodMethodNoReturn(param1 string, param2 string) {
	// do nothing
}

func (ss *simpleStruct) BadMethod(param1 complex64) complex64 {
	return param1
}

func (ss *simpleStruct) BadMethodGoodTransaction(ctx *TransactionContext, param1 complex64) complex64 {
	return param1
}

func (ss *simpleStruct) BadTransactionMethod(param1 string, ctx *TransactionContext) string {
	return param1
}

func (ss *simpleStruct) BadTransactionInterfaceMethod(ctx BadTransactionInterface, param1 string) string {
	return param1
}

func (ss *simpleStruct) BadReturnMethod(param1 string) (string, string, error) {
	return param1, "", nil
}

func (ss *simpleStruct) BadMethodFirstReturn(param1 complex64) (complex64, error) {
	return param1, nil
}

func (ss *simpleStruct) BadMethodSecondReturn(param1 string) (string, string) {
	return param1, param1
}

type mockSerializer struct{}

func (ms *mockSerializer) FromString(string, reflect.Type, *metadata.ParameterMetadata, *metadata.ComponentMetadata) (reflect.Value, error) {
	return reflect.ValueOf("HELLO WORLD"), nil
}

func (ms *mockSerializer) ToString(reflect.Value, reflect.Type, *metadata.ReturnMetadata, *metadata.ComponentMetadata) (string, error) {
	return "", errors.New("Serializer error")
}

func getMethodByName(strct interface{}, methodName string) (reflect.Method, reflect.Value) {
	strctType := reflect.TypeOf(strct)
	strctVal := reflect.ValueOf(strct)

	for i := 0; i < strctType.NumMethod(); i++ {
		if strctType.Method(i).Name == methodName {
			return strctType.Method(i), strctVal.Method(i)
		}
	}

	panic(fmt.Sprintf("Function with name %s does not exist for interface passed", methodName))
}

func setContractFunctionReturns(cf *ContractFunction, successReturn reflect.Type, returnsError bool) {
	cfr := contractFunctionReturns{}
	cfr.success = successReturn
	cfr.error = returnsError

	cf.returns = cfr
}

func testHandleResponse(t *testing.T, successReturn reflect.Type, errorReturn bool, response []reflect.Value, expectedString string, expectedValue interface{}, expectedError error, serializer serializer.TransactionSerializer) {
	t.Helper()

	cf := ContractFunction{}

	setContractFunctionReturns(&cf, successReturn, errorReturn)
	strResp, valueResp, errResp := cf.handleResponse(response, nil, nil, serializer)

	assert.Equal(t, expectedString, strResp, "should have returned string value from response")
	assert.Equal(t, expectedValue, valueResp, "should have returned actual value from response")
	assert.Equal(t, expectedError, errResp, "should have returned error value from response")
}

func createGoJSONSchemaSchema(propName string, schema *spec.Schema, components *metadata.ComponentMetadata) *gojsonschema.Schema {
	combined := make(map[string]interface{})
	combined["components"] = components
	combined["properties"] = make(map[string]interface{})
	combined["properties"].(map[string]interface{})[propName] = schema

	combinedLoader := gojsonschema.NewGoLoader(combined)

	gjs, _ := gojsonschema.NewSchema(combinedLoader)

	return gjs
}

// ================================
// Tests
// ================================

func TestHandleResponse(t *testing.T) {
	var response []reflect.Value
	err := errors.New("some error")

	serializer := new(serializer.JSONSerializer)

	// Should return error when wrong return length
	testHandleResponse(t, reflect.TypeOf(""), true, response, "", nil, errors.New("response does not match expected return for given function"), serializer)

	// Should return blank string and nil for error when no return specified
	testHandleResponse(t, nil, false, response, "", nil, nil, serializer)

	// Should return specified value for single success return
	response = []reflect.Value{reflect.ValueOf(1)}
	testHandleResponse(t, reflect.TypeOf(1), false, response, "1", 1, nil, serializer)

	// should return nil for error for single error return type when response is nil
	response = []reflect.Value{reflect.ValueOf(nil)}
	testHandleResponse(t, nil, true, response, "", nil, nil, serializer)

	// should return value for error for single error return type when response is an error
	response = []reflect.Value{reflect.ValueOf(err)}
	testHandleResponse(t, nil, true, response, "", nil, err, serializer)

	// should return nil for error and value for success for both success and error return type when response has nil error but success
	response = []reflect.Value{reflect.ValueOf(uint(1)), reflect.ValueOf(nil)}
	testHandleResponse(t, reflect.TypeOf(uint(1)), true, response, "1", uint(1), nil, serializer)

	// should return value for both error and success when function has both success and error return type and response has an error and success value
	response = []reflect.Value{reflect.ValueOf(true), reflect.ValueOf(err)}
	testHandleResponse(t, reflect.TypeOf(true), true, response, "true", true, err, serializer)

	// should return error when serializer ToString fails
	mockSerializerVal := new(mockSerializer)
	_, expectedErr := mockSerializerVal.ToString(reflect.ValueOf("NaN"), reflect.TypeOf(1), nil, nil)
	response = []reflect.Value{reflect.ValueOf("NaN")}
	testHandleResponse(t, reflect.TypeOf(1), false, response, "", nil, fmt.Errorf("Error handling success response. %s", expectedErr.Error()), mockSerializerVal)
}

func TestFormatArgs(t *testing.T) {
	var args []reflect.Value
	var err error

	fn := ContractFunction{}
	fn.params = contractFunctionParams{}
	fn.params.fields = []reflect.Type{reflect.TypeOf(1), reflect.TypeOf(2)}

	supplementaryMetadata := metadata.TransactionMetadata{}
	serializer := new(serializer.JSONSerializer)

	ctx := reflect.Value{}

	supplementaryMetadata.Parameters = []metadata.ParameterMetadata{}
	args, err = fn.formatArgs(ctx, supplementaryMetadata.Parameters, nil, []string{}, serializer)
	assert.EqualError(t, err, "Incorrect number of params in supplementary metadata. Expected 2, received 0", "should return error when metadata is incorrect")
	assert.Nil(t, args, "should not return values when metadata error occurs")

	args, err = fn.formatArgs(ctx, nil, nil, []string{}, serializer)
	assert.EqualError(t, err, "Incorrect number of params. Expected 2, received 0", "should return error when number of params is incorrect")
	assert.Nil(t, args, "should not return values when param error occurs")

	_, fromStringErr := serializer.FromString("NaN", reflect.TypeOf(1), nil, nil)
	args, err = fn.formatArgs(ctx, nil, nil, []string{"1", "NaN"}, serializer)
	assert.EqualError(t, err, fmt.Sprintf("Error managing parameter. %s", fromStringErr.Error()), "should return error when type of params is incorrect")
	assert.Nil(t, args, "should not return values when from string error occurs")

	args, err = fn.formatArgs(ctx, nil, nil, []string{"1", "2"}, serializer)
	assert.Nil(t, err, "should not error for valid values")
	assert.Equal(t, 1, args[0].Interface(), "should return converted values")
	assert.Equal(t, 2, args[1].Interface(), "should return converted values")

	supplementaryMetadata.Parameters = []metadata.ParameterMetadata{
		{
			Name:           "param1",
			Schema:         spec.Int64Property(),
			CompiledSchema: createGoJSONSchemaSchema("param1", spec.Int64Property(), nil),
		},
		{
			Name:           "param2",
			Schema:         spec.Int64Property(),
			CompiledSchema: createGoJSONSchemaSchema("param1", spec.Int64Property(), nil),
		},
	}
	args, err = fn.formatArgs(ctx, supplementaryMetadata.Parameters, nil, []string{"1", "2"}, serializer)
	assert.Nil(t, err, "should not error for valid values which validates against metadata")
	assert.Equal(t, 1, args[0].Interface(), "should return converted values validated against metadata")
	assert.Equal(t, 2, args[1].Interface(), "should return converted values validated against metadata")

	fn.params.context = reflect.TypeOf(ctx)
	args, err = fn.formatArgs(ctx, nil, nil, []string{"1", "2"}, serializer)
	assert.Nil(t, err, "should not error for valid values with context")
	assert.Equal(t, ctx, args[0], "should return converted values and context")
	assert.Equal(t, 1, args[1].Interface(), "should return converted values and context")
	assert.Equal(t, 2, args[2].Interface(), "should return converted values and context")
}

func TestMethodToContractFunctionParams(t *testing.T) {
	var params contractFunctionParams
	var err error
	var validTypeErr error

	ctx := reflect.TypeOf(new(TransactionContext))

	badMethod, _ := getMethodByName(new(simpleStruct), "BadMethod")
	validTypeErr = typeIsValid(reflect.TypeOf(complex64(1)), []reflect.Type{}, false)
	params, err = methodToContractFunctionParams(badMethod, ctx)
	assert.EqualError(t, err, fmt.Sprintf("BadMethod contains invalid parameter type. %s", validTypeErr.Error()), "should error when type is valid fails on first param")
	assert.Equal(t, params, contractFunctionParams{}, "should return blank params for invalid first param type")

	interfaceType := reflect.TypeOf((*BadTransactionInterface)(nil)).Elem()
	badInterfaceMethod, _ := getMethodByName(new(simpleStruct), "BadTransactionInterfaceMethod")
	matchesInterfaceErr := typeMatchesInterface(ctx, interfaceType)
	params, err = methodToContractFunctionParams(badInterfaceMethod, ctx)
	assert.EqualError(t, err, fmt.Sprintf("BadTransactionInterfaceMethod contains invalid transaction context interface type. Set transaction context for contract does not meet interface used in method. %s", matchesInterfaceErr.Error()), "should error when match on interface fails on first param")
	assert.Equal(t, params, contractFunctionParams{}, "should return blank params for invalid first param type")

	badCtxMethod, _ := getMethodByName(new(simpleStruct), "BadTransactionMethod")
	params, err = methodToContractFunctionParams(badCtxMethod, ctx)
	assert.EqualError(t, err, "Functions requiring the TransactionContext must require it as the first parameter. BadTransactionMethod takes it in as parameter 1", "should error when ctx in wrong position")
	assert.Equal(t, params, contractFunctionParams{}, "should return blank params when context in wrong position")

	badMethodGoodTransaction, _ := getMethodByName(new(simpleStruct), "BadMethodGoodTransaction")
	validTypeErr = typeIsValid(reflect.TypeOf(complex64(1)), []reflect.Type{}, false)
	params, err = methodToContractFunctionParams(badMethodGoodTransaction, ctx)
	assert.EqualError(t, err, fmt.Sprintf("BadMethodGoodTransaction contains invalid parameter type. %s", validTypeErr.Error()), "should error when type is valid fails but first param valid")
	assert.Equal(t, params, contractFunctionParams{}, "should return blank params for invalid param type when first param is ctx")

	goodMethod, _ := getMethodByName(new(simpleStruct), "GoodMethod")
	params, err = methodToContractFunctionParams(goodMethod, ctx)
	assert.Nil(t, err, "should not error for valid function")
	assert.Equal(t, params, contractFunctionParams{
		context: nil,
		fields: []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(""),
		},
	}, "should return params without context when none specified")

	goodTransactionMethod, _ := getMethodByName(new(simpleStruct), "GoodTransactionMethod")
	params, err = methodToContractFunctionParams(goodTransactionMethod, ctx)
	assert.Nil(t, err, "should not error for valid function")
	assert.Equal(t, params, contractFunctionParams{
		context: ctx,
		fields: []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(""),
		},
	}, "should return params with context when one specified")

	goodTransactionInterfaceMethod, _ := getMethodByName(new(simpleStruct), "GoodTransactionInterfaceMethod")
	params, err = methodToContractFunctionParams(goodTransactionInterfaceMethod, ctx)
	assert.Nil(t, err, "should not error for valid function")
	assert.Equal(t, params, contractFunctionParams{
		context: ctx,
		fields: []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(""),
		},
	}, "should return params with context when one specified")

	method := new(simpleStruct).GoodMethod
	funcMethod := reflect.Method{}
	funcMethod.Func = reflect.ValueOf(method)
	funcMethod.Type = reflect.TypeOf(method)
	params, err = methodToContractFunctionParams(funcMethod, ctx)
	assert.Nil(t, err, "should not error for valid function")
	assert.Equal(t, params, contractFunctionParams{
		context: nil,
		fields: []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(""),
		},
	}, "should return params without context when none specified for method from function")
}

func TestMethodToContractFunctionReturns(t *testing.T) {
	var returns contractFunctionReturns
	var err error
	var invalidTypeError error

	badReturnMethod, _ := getMethodByName(new(simpleStruct), "BadReturnMethod")
	returns, err = methodToContractFunctionReturns(badReturnMethod)
	assert.EqualError(t, err, "Functions may only return a maximum of two values. BadReturnMethod returns 3", "should error when more than two return values")
	assert.Equal(t, returns, contractFunctionReturns{}, "should return nothing for returns when errors for bad return length")

	badMethod, _ := getMethodByName(new(simpleStruct), "BadMethod")
	invalidTypeError = typeIsValid(reflect.TypeOf(complex64(1)), []reflect.Type{reflect.TypeOf((*error)(nil)).Elem()}, true)
	returns, err = methodToContractFunctionReturns(badMethod)
	assert.EqualError(t, err, fmt.Sprintf("BadMethod contains invalid single return type. %s", invalidTypeError.Error()), "should error when bad return type on single return")
	assert.Equal(t, returns, contractFunctionReturns{}, "should return nothing for returns when errors for single return type")

	badMethodFirstReturn, _ := getMethodByName(new(simpleStruct), "BadMethodFirstReturn")
	invalidTypeError = typeIsValid(reflect.TypeOf(complex64(1)), []reflect.Type{}, true)
	returns, err = methodToContractFunctionReturns(badMethodFirstReturn)
	assert.EqualError(t, err, fmt.Sprintf("BadMethodFirstReturn contains invalid first return type. %s", invalidTypeError.Error()), "should error when bad return type on first return")
	assert.Equal(t, returns, contractFunctionReturns{}, "should return nothing for returns when errors for first return type")

	badMethodSecondReturn, _ := getMethodByName(new(simpleStruct), "BadMethodSecondReturn")
	returns, err = methodToContractFunctionReturns(badMethodSecondReturn)
	assert.EqualError(t, err, "BadMethodSecondReturn contains invalid second return type. Type string is not valid. Expected error", "should error when bad return type on second return")
	assert.Equal(t, returns, contractFunctionReturns{}, "should return nothing for returns when errors for second return type")

	goodMethodNoReturn, _ := getMethodByName(new(simpleStruct), "GoodMethodNoReturn")
	returns, err = methodToContractFunctionReturns(goodMethodNoReturn)
	assert.Nil(t, err, "should not error when no return specified")
	assert.Equal(t, returns, contractFunctionReturns{nil, false}, "should return contractFunctionReturns for no return types")

	goodMethod, _ := getMethodByName(new(simpleStruct), "GoodMethod")
	returns, err = methodToContractFunctionReturns(goodMethod)
	assert.Nil(t, err, "should not error when single non error return type specified")
	assert.Equal(t, returns, contractFunctionReturns{reflect.TypeOf(""), false}, "should return contractFunctionReturns for single error return types")

	goodErrorMethod, _ := getMethodByName(new(simpleStruct), "GoodErrorMethod")
	returns, err = methodToContractFunctionReturns(goodErrorMethod)
	assert.Nil(t, err, "should not error when single error return type specified")
	assert.Equal(t, returns, contractFunctionReturns{nil, true}, "should return contractFunctionReturns for single error return types")

	goodReturnMethod, _ := getMethodByName(new(simpleStruct), "GoodReturnMethod")
	returns, err = methodToContractFunctionReturns(goodReturnMethod)
	assert.Nil(t, err, "should not error when good double return type specified")
	assert.Equal(t, returns, contractFunctionReturns{reflect.TypeOf(""), true}, "should return contractFunctionReturns for double return types")

	method := new(simpleStruct).GoodReturnMethod
	funcMethod := reflect.Method{}
	funcMethod.Func = reflect.ValueOf(method)
	funcMethod.Type = reflect.TypeOf(method)
	returns, err = methodToContractFunctionReturns(funcMethod)
	assert.Nil(t, err, "should not error when good double return type specified when method got from function")
	assert.Equal(t, returns, contractFunctionReturns{reflect.TypeOf(""), true}, "should return contractFunctionReturns for double return types when method got from function")
}

func TestParseMethod(t *testing.T) {
	var params contractFunctionParams
	var returns contractFunctionReturns
	var err error

	ctx := reflect.TypeOf(new(TransactionContext))

	badMethod, _ := getMethodByName(new(simpleStruct), "BadMethod")
	_, paramErr := methodToContractFunctionParams(badMethod, ctx)
	params, returns, err = parseMethod(badMethod, ctx)
	assert.EqualError(t, err, paramErr.Error(), "should return an error when get params errors")
	assert.Equal(t, contractFunctionParams{}, params, "should return no param detail when get params errors")
	assert.Equal(t, contractFunctionReturns{}, returns, "should return no return detail when get params errors")

	badReturnMethod, _ := getMethodByName(new(simpleStruct), "BadReturnMethod")
	_, returnErr := methodToContractFunctionReturns(badReturnMethod)
	params, returns, err = parseMethod(badReturnMethod, ctx)
	assert.EqualError(t, err, returnErr.Error(), "should return an error when get returns errors")
	assert.Equal(t, contractFunctionParams{}, params, "should return no param detail when get returns errors")
	assert.Equal(t, contractFunctionReturns{}, returns, "should return no return detail when get returns errors")

	goodMethod, _ := getMethodByName(new(simpleStruct), "GoodMethod")
	expectedParam, _ := methodToContractFunctionParams(goodMethod, ctx)
	expectedReturn, _ := methodToContractFunctionReturns(goodMethod)
	params, returns, err = parseMethod(goodMethod, ctx)
	assert.Nil(t, err, "should not error for valid function")
	assert.Equal(t, expectedParam, params, "should return params for valid function")
	assert.Equal(t, expectedReturn, returns, "should return returns for valid function")
}

func TestNewContractFunction(t *testing.T) {
	method := new(simpleStruct).GoodMethod
	fnValue := reflect.ValueOf(method)

	params := contractFunctionParams{
		nil,
		[]reflect.Type{reflect.TypeOf("")},
	}

	returns := contractFunctionReturns{
		reflect.TypeOf(""),
		true,
	}

	expectedCf := &ContractFunction{fnValue, CallTypeEvaluate, params, returns}

	cf := newContractFunction(fnValue, CallTypeEvaluate, params, returns)

	assert.Equal(t, cf, expectedCf, "should create contract function from passed in components")
}

func TestNewContractFunctionFromFunc(t *testing.T) {
	var cf *ContractFunction
	var err error
	var method interface{}
	var funcMethod reflect.Method

	ctx := reflect.TypeOf(new(TransactionContext))

	cf, err = NewContractFunctionFromFunc("", CallTypeSubmit, ctx)
	assert.EqualError(t, err, "Cannot create new contract function from string. Can only use func", "should return error if interface passed not a func")
	assert.Nil(t, cf, "should not return contract function if interface passed not a func")

	method = new(simpleStruct).BadMethod
	funcMethod = reflect.Method{}
	funcMethod.Func = reflect.ValueOf(method)
	funcMethod.Type = reflect.TypeOf(method)
	_, _, parseErr := parseMethod(funcMethod, ctx)
	cf, err = NewContractFunctionFromFunc(method, CallTypeSubmit, ctx)
	assert.EqualError(t, err, parseErr.Error(), "should return error from failed parsing")
	assert.Nil(t, cf, "should not return contract function if parse fails")

	method = new(simpleStruct).GoodMethod
	funcMethod = reflect.Method{}
	funcMethod.Func = reflect.ValueOf(method)
	funcMethod.Type = reflect.TypeOf(method)
	params, returns, _ := parseMethod(funcMethod, ctx)
	expectedCf := newContractFunction(reflect.ValueOf(method), CallTypeSubmit, params, returns)
	cf, err = NewContractFunctionFromFunc(method, CallTypeSubmit, ctx)
	assert.Nil(t, err, "should not error when parse successful from func")
	assert.Equal(t, expectedCf, cf, "should return contract function for good method from func")
}

func TestNewContractFunctionFromReflect(t *testing.T) {
	var cf *ContractFunction
	var err error

	ctx := reflect.TypeOf(new(TransactionContext))

	badMethod, badMethodValue := getMethodByName(new(simpleStruct), "BadMethod")
	_, _, parseErr := parseMethod(badMethod, ctx)
	cf, err = NewContractFunctionFromReflect(badMethod, badMethodValue, CallTypeEvaluate, ctx)
	assert.EqualError(t, err, parseErr.Error(), "should return parse error on parsing failure")
	assert.Nil(t, cf, "should not return contract function on error")

	goodMethod, goodMethodValue := getMethodByName(new(simpleStruct), "GoodMethod")
	params, returns, _ := parseMethod(goodMethod, ctx)
	expectedCf := newContractFunction(goodMethodValue, CallTypeEvaluate, params, returns)
	cf, err = NewContractFunctionFromReflect(goodMethod, goodMethodValue, CallTypeEvaluate, ctx)
	assert.Nil(t, err, "should not error when parse successful from reflect")
	assert.Equal(t, expectedCf, cf, "should return contract function for good method from reflect")
}

func TestReflectMetadata(t *testing.T) {
	var txMetadata metadata.TransactionMetadata
	var testCf ContractFunction

	testCf = ContractFunction{
		params: contractFunctionParams{
			nil,
			[]reflect.Type{reflect.TypeOf(""), reflect.TypeOf(true)},
		},
		returns: contractFunctionReturns{
			success: reflect.TypeOf(1),
		},
	}

	txMetadata = testCf.ReflectMetadata("some tx", nil)
	expectedMetadata := metadata.TransactionMetadata{
		Parameters: []metadata.ParameterMetadata{
			{Name: "param0", Schema: spec.StringProperty()},
			{Name: "param1", Schema: spec.BoolProperty()},
		},
		Returns: metadata.ReturnMetadata{Schema: spec.Int64Property()},
		Tag:     []string{"submit"},
		Name:    "some tx",
	}
	assert.Equal(t, expectedMetadata, txMetadata, "should return metadata for submit transaction")

	testCf.callType = CallTypeEvaluate
	txMetadata = testCf.ReflectMetadata("some tx", nil)
	expectedMetadata.Tag = []string{"evaluate"}
	assert.Equal(t, expectedMetadata, txMetadata, "should return metadata for evaluate transaction")
}

func TestCall(t *testing.T) {
	var expectedStr string
	var expectedIface interface{}
	var expectedErr error
	var actualStr string
	var actualIface interface{}
	var actualErr error

	ctx := reflect.ValueOf(TransactionContext{})

	testCf := ContractFunction{
		function: reflect.ValueOf(new(simpleStruct).GoodMethod),
		params: contractFunctionParams{
			nil,
			[]reflect.Type{reflect.TypeOf(""), reflect.TypeOf("")},
		},
		returns: contractFunctionReturns{
			success: reflect.TypeOf(""),
		},
	}

	serializer := new(serializer.JSONSerializer)

	actualStr, actualIface, actualErr = testCf.Call(ctx, nil, nil, serializer, "some data")
	_, expectedErr = testCf.formatArgs(ctx, nil, nil, []string{"some data"}, serializer)
	assert.EqualError(t, actualErr, expectedErr.Error(), "should error when formatting args fails")
	assert.Nil(t, actualIface, "should not return an interface when format args fails")
	assert.Equal(t, "", actualStr, "should return empty string when format args fails")

	expectedStr, expectedIface, expectedErr = testCf.handleResponse([]reflect.Value{reflect.ValueOf("helloworld")}, nil, nil, serializer)
	actualStr, actualIface, actualErr = testCf.Call(ctx, nil, nil, serializer, "hello", "world")
	assert.Equal(t, actualErr, expectedErr, "should return same error as handle response for good function")
	assert.Equal(t, expectedStr, actualStr, "should return same string as handle response for good function and params")
	assert.Equal(t, expectedIface, expectedIface, "should return same interface as handle response for good function and params")

	combined := make(map[string]interface{})
	combined["components"] = nil
	combined["properties"] = make(map[string]interface{})
	combined["properties"].(map[string]interface{})["param0"] = spec.StringProperty()
	combined["properties"].(map[string]interface{})["param1"] = spec.StringProperty()
	combined["properties"].(map[string]interface{})["returns"] = spec.StringProperty()
	combinedLoader := gojsonschema.NewGoLoader(combined)
	compiled, _ := gojsonschema.NewSchema(combinedLoader)
	schema := metadata.TransactionMetadata{}
	schema.Parameters = []metadata.ParameterMetadata{
		{Name: "param0", Schema: spec.StringProperty(), CompiledSchema: compiled},
		{Name: "param1", Schema: spec.StringProperty(), CompiledSchema: compiled},
	}
	schema.Returns = metadata.ReturnMetadata{Schema: spec.StringProperty(), CompiledSchema: compiled}
	expectedStr, expectedIface, expectedErr = testCf.handleResponse([]reflect.Value{reflect.ValueOf("helloworld")}, &schema.Returns, nil, serializer)
	actualStr, actualIface, actualErr = testCf.Call(ctx, &schema, nil, serializer, "hello", "world")
	assert.Equal(t, actualErr, expectedErr, "should return same error as handle response for good function with schema")
	assert.Equal(t, expectedStr, actualStr, "should return same string as handle response for good function and params with schema")
	assert.Equal(t, expectedIface, expectedIface, "should return same interface as handle response for good function and params with schema")
}
