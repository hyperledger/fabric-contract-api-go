// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package metadata

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		return os.ReadFile(filename)
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
	expectedSchema, err := readLocalFile("schema/schema.json")
	require.NoError(t, err, "read schema file")

	schema := GetJSONSchema()

	assert.Equal(t, expectedSchema, schema, "should return same schema as in file")
}

func TestUnmarshalJSON(t *testing.T) {
	ttm := new(TransactionMetadata)

	err := json.Unmarshal([]byte("{\"name\": 1}"), ttm)
	require.Error(t, err, "should error on bad JSON")
	assert.Regexpf(t, "json: cannot unmarshal number into Go struct field .*\\.name of type string", err.Error(), "should error on bad JSON")

	err = json.Unmarshal([]byte("{\"name\":\"Transaction1\",\"returns\":{\"type\":\"string\"}}"), ttm)
	require.NoError(t, err, "should not error on valid json")
	assert.Equal(t, &TransactionMetadata{Name: "Transaction1", Returns: ReturnMetadata{Schema: spec.StringProperty()}}, ttm, "should setup TransactionMetadata from json bytes")

}

func TestMarshalJSON(t *testing.T) {
	ttm := TransactionMetadata{Name: "Transaction1", Returns: ReturnMetadata{Schema: spec.StringProperty()}}
	bytes, err := json.Marshal(&ttm)

	require.NoError(t, err, "should not error on marshall")
	assert.JSONEqf(t, "{\"name\":\"Transaction1\",\"returns\":{\"type\":\"string\"}}", string(bytes), "should return JSON with returns as schema not object")
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
		Schema: spec.RefProperty("non-existent"),
	}

	badParameter := ParameterMetadata{
		Name:   "badParam",
		Schema: spec.RefProperty("non-existent"),
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
	assert.Contains(t, err.Error(), "error compiling schema for someContract [someTransaction]. Return schema invalid.", "should error on bad schema for return value")

	someTransaction.Parameters = []ParameterMetadata{badParameter}
	someContract.Transactions[0] = someTransaction
	ccm.Contracts["someContract"] = someContract
	err = ccm.CompileSchemas()
	assert.Contains(t, err.Error(), "error compiling schema for someContract [someTransaction]. badParam schema invalid.", "should error on bad schema for param value")

	someTransaction.Returns = goodReturn
	someTransaction.Parameters = []ParameterMetadata{goodParameter1, goodParameter2}
	someContract.Transactions[0] = someTransaction
	ccm.Contracts["someContract"] = someContract
	err = ccm.CompileSchemas()
	require.NoError(t, err, "should not error on good metadata")
	validateCompiledSchema(t, "goodParam1", make(map[string]interface{}), ccm.Contracts["someContract"].Transactions[0].Parameters[0].CompiledSchema)
	validateCompiledSchema(t, "goodParam2", "abc", ccm.Contracts["someContract"].Transactions[0].Parameters[1].CompiledSchema)
	validateCompiledSchema(t, "return", 1, ccm.Contracts["someContract"].Transactions[0].Returns.CompiledSchema)

	someTransaction.Returns = nilReturn
	someTransaction.Parameters = []ParameterMetadata{goodParameter1, goodParameter2}
	someContract.Transactions[0] = someTransaction
	ccm.Contracts["someContract"] = someContract
	err = ccm.CompileSchemas()
	require.NoError(t, err, "should not error on good metadata when return is nil")
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

	assert.True(t, result.Valid(), "should validate for %s compiled schema", propName)
}

func TestReadMetadataFile(t *testing.T) {
	expectedContractChaincodeMetadata := ContractChaincodeMetadata{}
	metadataBytes := []byte("{\"info\":{\"title\":\"my contract\",\"version\":\"0.0.1\"},\"contracts\":{},\"components\":{}}")
	require.NoError(t, json.Unmarshal(metadataBytes, &expectedContractChaincodeMetadata))

	t.Run("Exec error reading metadata file", func(t *testing.T) {
		fakeOS(t, osExcTestStr{})

		metadata, err := ReadMetadataFile()
		require.EqualError(t, err, "failed to read metadata from file. Could not find location of executable. some error", "should error when cannot read file due to exec error")
		assert.Equal(t, ContractChaincodeMetadata{}, metadata, "should return blank metadata when cannot read file due to exec error")
	})

	t.Run("Stat error reading metadata file", func(t *testing.T) {
		fakeOS(t, osStatTestStr{})

		metadata, err := ReadMetadataFile()
		require.EqualError(t, err, "failed to read metadata from file. Metadata file does not exist", "should error when cannot read file due to stat error")
		assert.Equal(t, ContractChaincodeMetadata{}, metadata, "should return blank metadata when cannot read file due to stat error")
	})

	t.Run("Error reading metadata file from contract-metadata directory", func(t *testing.T) {
		fakeOS(t, osStatTestStrContractMeta{})
		ContractMetaNumberOfCalls = 0

		metadata, err := ReadMetadataFile()
		assert.Equal(t, 2, ContractMetaNumberOfCalls, "Should check contract-metadata directory if META-INF doesn't contain metadata.json file")
		assert.Contains(t, err.Error(), "failed to read metadata from file. Could not read file", "should error when cannot read file due to read error")
		assert.Equal(t, ContractChaincodeMetadata{}, metadata, "should return blank metadata when cannot read file due to read error")
	})

	t.Run("Returns blank metadata on file read error", func(t *testing.T) {
		fakeOS(t, osWorkTestStr{})
		fakeIOUtil(t, ioUtilReadFileTestStr{})

		metadata, err := ReadMetadataFile()
		assert.Contains(t, err.Error(), "failed to read metadata from file. Could not read file", "should error when cannot read file due to read error")
		assert.Equal(t, ContractChaincodeMetadata{}, metadata, "should return blank metadata when cannot read file due to read error")
	})

	t.Run("Returns contract metadata from META-INF", func(t *testing.T) {
		fakeOS(t, osWorkTestStr{})
		fakeIOUtil(t, ioUtilWorkTestStr{})

		metadata, err := ReadMetadataFile()
		require.NoError(t, err, "should not return error when can read file")

		assert.Equal(t, expectedContractChaincodeMetadata, metadata, "should return contract metadata that was in the file")
	})

	t.Run("Returns contract metadata from contract-metadata", func(t *testing.T) {
		fakeOS(t, osWorkTestStrContractMeta{})
		fakeIOUtil(t, ioUtilWorkTestStr{})
		ContractMetaNumberOfCalls = 0

		metadata, err := ReadMetadataFile()
		assert.Equal(t, 2, ContractMetaNumberOfCalls, "Should check contract-metadata directory if META-INF doesn't contain metadata.json file")
		require.NoError(t, err, "should not return error when can read file")
		assert.Equal(t, expectedContractChaincodeMetadata, metadata, "should return contract metadata that was in the file")
	})
}

func TestValidateAgainstSchema(t *testing.T) {
	fakeOS(t, osWorkTestStr{})
	fakeIOUtil(t, ioUtilWorkTestStr{})

	t.Run("Error on empty metadata", func(t *testing.T) {
		err := ValidateAgainstSchema(ContractChaincodeMetadata{})
		require.EqualError(t, err, "cannot use metadata. Metadata did not match schema:\n1. (root): info is required\n2. contracts: Invalid type. Expected: object, given: null", "should error when metadata given does not match schema")
	})

	t.Run("Valid metadata", func(t *testing.T) {
		metadata, err := ReadMetadataFile()
		require.NoError(t, err)

		err = ValidateAgainstSchema(metadata)
		require.NoError(t, err, "should not error for valid metadata")
	})
}

func fakeOS(t *testing.T, fake osInterface) {
	previous := osAbs
	t.Cleanup(func() {
		osAbs = previous
	})
	osAbs = fake
}

func fakeIOUtil(t *testing.T, fake ioutilInterface) {
	previous := ioutilAbs
	t.Cleanup(func() {
		ioutilAbs = previous
	})
	ioutilAbs = fake
}
