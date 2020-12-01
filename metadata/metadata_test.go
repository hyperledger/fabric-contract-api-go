// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

// ================================
// Helpers
// ================================

var ContractMetaNumberOfCalls int

type ioUtilReadFileTestStr struct{}

func (io ioUtilReadFileTestStr) ReadFile(filename string) ([]byte, error) {
	return nil, errors.New("some error")
}

type ioUtilWorkTestStr struct{}

func (io ioUtilWorkTestStr) ReadFile(filename string) ([]byte, error) {
	if strings.Contains(filename, "schema.json") {
		return ioutil.ReadFile(filename)
	}

	return []byte("{\"info\":{\"title\":\"my contract\",\"version\":\"0.0.1\"},\"contracts\":{},\"components\":{}}"), nil
}

type osExcTestStr struct{}

func (o osExcTestStr) Executable() (string, error) {
	return "", errors.New("some error")
}

func (o osExcTestStr) Stat(name string) (os.FileInfo, error) {
	return nil, nil
}

func (o osExcTestStr) IsNotExist(err error) bool {
	return false
}

type osStatTestStr struct{}

func (o osStatTestStr) Executable() (string, error) {
	return "", nil
}

func (o osStatTestStr) Stat(name string) (os.FileInfo, error) {
	return os.Stat("some bad file")
}

func (o osStatTestStr) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

type osStatTestStrContractMeta struct{}

func (o osStatTestStrContractMeta) Executable() (string, error) {
	return "", nil
}

func (o osStatTestStrContractMeta) Stat(name string) (os.FileInfo, error) {
	ContractMetaNumberOfCalls++
	if ContractMetaNumberOfCalls == 1 {
		ContractMetaNumberOfCalls++
		return os.Stat("some bad file")
	}
	return os.Stat("some good file")
}

func (o osStatTestStrContractMeta) IsNotExist(err error) bool {
	return false
}

type osWorkTestStrContractMeta struct{}

func (o osWorkTestStrContractMeta) Executable() (string, error) {
	return "", nil
}

func (o osWorkTestStrContractMeta) Stat(name string) (os.FileInfo, error) {
	ContractMetaNumberOfCalls++
	if ContractMetaNumberOfCalls == 1 {
		ContractMetaNumberOfCalls++
		return os.Stat("some bad file")
	}
	return os.Stat("some good file")
}

func (o osWorkTestStrContractMeta) IsNotExist(err error) bool {
	return false
}

type osWorkTestStr struct{}

func (o osWorkTestStr) Executable() (string, error) {
	return "", nil
}

func (o osWorkTestStr) Stat(name string) (os.FileInfo, error) {
	return os.Stat("some good file")
}

func (o osWorkTestStr) IsNotExist(err error) bool {
	return false
}

// ================================
// Tests
// ================================

func TestGetJSONSchema(t *testing.T) {
	var schema []byte
	var err error

	expectedSchema, expectedErr := readLocalFile("schema/schema.json")
	schema, err = GetJSONSchema()

	if expectedErr != nil {
		panic("TEST FAILED. Reading schema should not return error")
	}

	assert.Nil(t, err, "should not error when getting schema")
	assert.Equal(t, expectedSchema, schema, "should return same schema as in file. Have you updated schema without running packr?")
}

func TestUnmarshalJSON(t *testing.T) {
	ttm := new(TransactionMetadata)

	err := json.Unmarshal([]byte("{\"name\": 1}"), ttm)
	assert.EqualError(t, err, "json: cannot unmarshal number into Go struct field jsonTransactionMetadata.name of type string", "should error on bad JSON")

	err = json.Unmarshal([]byte("{\"name\":\"Transaction1\",\"returns\":{\"type\":\"string\"}}"), ttm)
	assert.Nil(t, err, "should not error on valid json")
	assert.Equal(t, &TransactionMetadata{Name: "Transaction1", Returns: ReturnMetadata{Schema: spec.StringProperty()}}, ttm, "should setup TransactionMetadata from json bytes")

}

func TestMarshalJSON(t *testing.T) {
	ttm := TransactionMetadata{Name: "Transaction1", Returns: ReturnMetadata{Schema: spec.StringProperty()}}
	bytes, err := json.Marshal(&ttm)

	assert.Nil(t, err, "should not error on marshall")
	assert.Equal(t, "{\"name\":\"Transaction1\",\"returns\":{\"type\":\"string\"}}", string(bytes), "should return JSON with returns as schema not object")
}

func TestAppend(t *testing.T) {
	var ccm ContractChaincodeMetadata

	source := ContractChaincodeMetadata{}
	source.Info = new(InfoMetadata)
	source.Info.Title = "A title"
	source.Info.Version = "Some version"

	someContract := ContractMetadata{}
	someContract.Name = "some contract"

	source.Contracts = make(map[string]ContractMetadata)
	source.Contracts["some contract"] = someContract

	someComponent := ObjectMetadata{}

	source.Components = ComponentMetadata{}
	source.Components.Schemas = make(map[string]ObjectMetadata)
	source.Components.Schemas["some component"] = someComponent

	// should use the source info when info is blank
	ccm = ContractChaincodeMetadata{}
	ccm.Append(source)

	assert.Equal(t, source.Info, ccm.Info, "should have used source info when info blank")

	// should use own info when info set
	ccm = ContractChaincodeMetadata{}
	ccm.Info = new(InfoMetadata)
	ccm.Info.Title = "An existing title"
	ccm.Info.Version = "Some existing version"

	someInfo := ccm.Info

	ccm.Append(source)

	assert.Equal(t, someInfo, ccm.Info, "should have used own info when info existing")
	assert.NotEqual(t, source.Info, ccm.Info, "should not use source info when info exists")

	// should use the source contract when contract is 0 length and nil
	ccm = ContractChaincodeMetadata{}
	ccm.Append(source)

	assert.Equal(t, source.Contracts, ccm.Contracts, "should have used source info when contract 0 length map")

	// should use the source contract when contract is 0 length and not nil
	ccm = ContractChaincodeMetadata{}
	ccm.Contracts = make(map[string]ContractMetadata)
	ccm.Append(source)

	assert.Equal(t, source.Contracts, ccm.Contracts, "should have used source info when contract 0 length map")

	// should use own contract when contract greater than 1
	anotherContract := ContractMetadata{}
	anotherContract.Name = "some contract"

	ccm = ContractChaincodeMetadata{}
	ccm.Contracts = make(map[string]ContractMetadata)
	ccm.Contracts["another contract"] = anotherContract

	contractMap := ccm.Contracts

	assert.Equal(t, contractMap, ccm.Contracts, "should have used own contracts when contracts existing")
	assert.NotEqual(t, source.Contracts, ccm.Contracts, "should not have used source contracts when existing contracts")

	// should use source components when components is empty
	ccm = ContractChaincodeMetadata{}
	ccm.Append(source)

	assert.Equal(t, ccm.Components, source.Components, "should use sources components")

	// should use own components when components is empty
	anotherComponent := ObjectMetadata{}

	ccm = ContractChaincodeMetadata{}
	ccm.Components = ComponentMetadata{}
	ccm.Components.Schemas = make(map[string]ObjectMetadata)
	ccm.Components.Schemas["another component"] = anotherComponent

	ccmComponent := ccm.Components

	ccm.Append(source)

	assert.Equal(t, ccmComponent, ccm.Components, "should have used own components")
	assert.NotEqual(t, source.Components, ccm.Components, "should not be same as source components")
}

func TestCompileSchemas(t *testing.T) {
	var err error

	badReturn := ReturnMetadata{
		Schema: spec.RefProperty("non-existant"),
	}

	badParameter := ParameterMetadata{
		Name:   "badParam",
		Schema: spec.RefProperty("non-existant"),
	}

	goodReturn := ReturnMetadata{
		Schema: spec.Int64Property(),
	}

	nilReturn := ReturnMetadata{
		Schema: nil,
	}

	goodParameter1 := ParameterMetadata{
		Name:   "goodParam1",
		Schema: spec.RefProperty("#/components/schemas/someComponent"),
	}

	goodParameter2 := ParameterMetadata{
		Name:   "goodParam2",
		Schema: spec.StringProperty(),
	}

	someComponent := ObjectMetadata{
		Properties: make(map[string]spec.Schema),
		Required:   []string{},
	}
	someTransaction := TransactionMetadata{
		Name: "someTransaction",
	}
	someContract := ContractMetadata{
		Transactions: []TransactionMetadata{someTransaction},
	}

	ccm := ContractChaincodeMetadata{}
	ccm.Components = ComponentMetadata{}
	ccm.Components.Schemas = make(map[string]ObjectMetadata)
	ccm.Components.Schemas["someComponent"] = someComponent
	ccm.Contracts = make(map[string]ContractMetadata)
	ccm.Contracts["someContract"] = someContract

	someTransaction.Returns = badReturn
	someContract.Transactions[0] = someTransaction
	ccm.Contracts["someContract"] = someContract
	err = ccm.CompileSchemas()
	assert.Contains(t, err.Error(), "Error compiling schema for someContract [someTransaction]. Return schema invalid.", "should error on bad schema for return value")

	someTransaction.Parameters = []ParameterMetadata{badParameter}
	someContract.Transactions[0] = someTransaction
	ccm.Contracts["someContract"] = someContract
	err = ccm.CompileSchemas()
	assert.Contains(t, err.Error(), "Error compiling schema for someContract [someTransaction]. badParam schema invalid.", "should error on bad schema for param value")

	someTransaction.Returns = goodReturn
	someTransaction.Parameters = []ParameterMetadata{goodParameter1, goodParameter2}
	someContract.Transactions[0] = someTransaction
	ccm.Contracts["someContract"] = someContract
	err = ccm.CompileSchemas()
	assert.Nil(t, err, "should not error on good metadata")
	validateCompiledSchema(t, "goodParam1", make(map[string]interface{}), ccm.Contracts["someContract"].Transactions[0].Parameters[0].CompiledSchema)
	validateCompiledSchema(t, "goodParam2", "abc", ccm.Contracts["someContract"].Transactions[0].Parameters[1].CompiledSchema)
	validateCompiledSchema(t, "return", 1, ccm.Contracts["someContract"].Transactions[0].Returns.CompiledSchema)

	someTransaction.Returns = nilReturn
	someTransaction.Parameters = []ParameterMetadata{goodParameter1, goodParameter2}
	someContract.Transactions[0] = someTransaction
	ccm.Contracts["someContract"] = someContract
	err = ccm.CompileSchemas()
	assert.Nil(t, err, "should not error on good metadata when return is nil")
	validateCompiledSchema(t, "goodParam1", make(map[string]interface{}), ccm.Contracts["someContract"].Transactions[0].Parameters[0].CompiledSchema)
	validateCompiledSchema(t, "goodParam2", "abc", ccm.Contracts["someContract"].Transactions[0].Parameters[1].CompiledSchema)
	assert.Nil(t, ccm.Contracts["someContract"].Transactions[0].Returns.CompiledSchema, "should set compiled schema nil on no return")
}

func validateCompiledSchema(t *testing.T, propName string, propValue interface{}, compiledSchema *gojsonschema.Schema) {
	t.Helper()

	returnValidator := make(map[string]interface{})
	returnValidator["return"] = propValue

	toValidateLoader := gojsonschema.NewGoLoader(returnValidator)

	result, _ := compiledSchema.Validate(toValidateLoader)

	assert.True(t, result.Valid(), fmt.Sprintf("should validate for %s compiled schema", propName))
}

func TestReadMetadataFile(t *testing.T) {
	ContractMetaNumberOfCalls = 0
	var metadata ContractChaincodeMetadata
	var err error

	oldOsHelper := osAbs

	osAbs = osExcTestStr{}
	metadata, err = ReadMetadataFile()
	assert.EqualError(t, err, "Failed to read metadata from file. Could not find location of executable. some error", "should error when cannot read file due to exec error")
	assert.Equal(t, ContractChaincodeMetadata{}, metadata, "should return blank metadata when cannot read file due to exec error")

	osAbs = osStatTestStr{}
	metadata, err = ReadMetadataFile()
	assert.EqualError(t, err, "Failed to read metadata from file. Metadata file does not exist", "should error when cannot read file due to stat error")
	assert.Equal(t, ContractChaincodeMetadata{}, metadata, "should return blank metadata when cannot read file due to stat error")

	osAbs = osStatTestStrContractMeta{}
	metadata, err = ReadMetadataFile()
	assert.Equal(t, ContractMetaNumberOfCalls, 2, "Should check contract-metadata directory if META-INF doesn't contain metadata.json file")
	assert.Contains(t, err.Error(), "Failed to read metadata from file. Could not read file", "should error when cannot read file due to read error")
	assert.Equal(t, ContractChaincodeMetadata{}, metadata, "should return blank metadata when cannot read file due to read error")
	ContractMetaNumberOfCalls = 0

	oldIoUtilHelper := ioutilAbs
	osAbs = osWorkTestStr{}

	ioutilAbs = ioUtilReadFileTestStr{}
	metadata, err = ReadMetadataFile()
	assert.Contains(t, err.Error(), "Failed to read metadata from file. Could not read file", "should error when cannot read file due to read error")
	assert.Equal(t, ContractChaincodeMetadata{}, metadata, "should return blank metadata when cannot read file due to read error")

	ioutilAbs = ioUtilWorkTestStr{}
	metadata, err = ReadMetadataFile()
	metadataBytes := []byte("{\"info\":{\"title\":\"my contract\",\"version\":\"0.0.1\"},\"contracts\":{},\"components\":{}}")
	expectedContractChaincodeMetadata := ContractChaincodeMetadata{}
	json.Unmarshal(metadataBytes, &expectedContractChaincodeMetadata)
	assert.Nil(t, err, "should not return error when can read file")
	assert.Equal(t, expectedContractChaincodeMetadata, metadata, "should return contract metadata that was in the file")

	osAbs = osWorkTestStrContractMeta{}
	metadata, err = ReadMetadataFile()
	assert.Equal(t, ContractMetaNumberOfCalls, 2, "Should check contract-metadata directory if META-INF doesn't contain metadata.json file")
	assert.Nil(t, err, "should not return error when can read file")
	assert.Equal(t, expectedContractChaincodeMetadata, metadata, "should return contract metadata that was in the file")
	ContractMetaNumberOfCalls = 0

	ioutilAbs = oldIoUtilHelper
	osAbs = oldOsHelper
}

func TestValidateAgainstSchema(t *testing.T) {
	var err error

	oldIoUtilHelper := ioutilAbs
	oldOsHelper := osAbs
	osAbs = osWorkTestStr{}

	metadata := ContractChaincodeMetadata{}

	ioutilAbs = ioUtilWorkTestStr{}

	err = ValidateAgainstSchema(metadata)
	assert.EqualError(t, err, "Cannot use metadata. Metadata did not match schema:\n1. (root): info is required\n2. contracts: Invalid type. Expected: object, given: null", "should error when metadata given does not match schema")

	metadata, _ = ReadMetadataFile()
	err = ValidateAgainstSchema(metadata)
	assert.Nil(t, err, "should not error for valid metadata")

	ioutilAbs = oldIoUtilHelper
	osAbs = oldOsHelper
}
