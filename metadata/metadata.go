// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package metadata

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/go-openapi/spec"
	"github.com/hyperledger/fabric-contract-api-go/v2/internal/schema"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// MetadataFolder name of the main folder metadata should be placed in
const MetadataFolder = "META-INF"

// MetadataFolderSecondary name of the secondary folder metadata should be placed in
const MetadataFolderSecondary = "contract-metadata"

// MetadataFile name of file metadata should be written in
const MetadataFile = "metadata.json"

// Helpers for testing
type osInterface interface {
	Getwd() (string, error)
	ReadFile(string) ([]byte, error)
	IsNotExist(error) bool
}

type osFront struct{}

func (o *osFront) Getwd() (string, error) {
	return os.Getwd()
}

func (o *osFront) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (o *osFront) IsNotExist(err error) bool {
	return errors.Is(err, fs.ErrNotExist)
}

var osAbs osInterface = &osFront{}

//go:embed schema/schema.json
var contractSchemaJSON []byte

// GetJSONSchema returns the JSON schema used for metadata
func GetJSONSchema() []byte {
	return contractSchemaJSON
}

var getCompiledSchema = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return schema.CompileRawSchema("schema.json", GetJSONSchema())
})

// SchemaValidator validates a property value against a compiled schema.
type SchemaValidator interface {
	Validate(propName string, typ reflect.Type, stringValue string, obj any) error
}

// ParameterMetadata details about a parameter used for a transaction.
type ParameterMetadata struct {
	Description string          `json:"description,omitempty"`
	Name        string          `json:"name"`
	Schema      *spec.Schema    `json:"schema"`
	Validator   SchemaValidator `json:"-"`
}

// ReturnMetadata details about the return type for a transaction
type ReturnMetadata struct {
	Schema    *spec.Schema
	Validator SchemaValidator `json:"-"`
}

// TransactionMetadata contains information on what makes up a transaction
// When JSON serialized the Returns object is flattened to contain the schema
type TransactionMetadata struct {
	Parameters []ParameterMetadata `json:"parameters,omitempty"`
	Returns    ReturnMetadata      `json:"-"`
	Tag        []string            `json:"tag,omitempty"`
	Name       string              `json:"name"`
}

type tmAlias TransactionMetadata
type jsonTransactionMetadata struct {
	*tmAlias
	ReturnsSchema *spec.Schema `json:"returns,omitempty"`
}

// UnmarshalJSON handles converting JSON to TransactionMetadata since returns is flattened
// in swagger
func (tm *TransactionMetadata) UnmarshalJSON(data []byte) error {
	jtm := jsonTransactionMetadata{tmAlias: (*tmAlias)(tm)}

	err := json.Unmarshal(data, &jtm)

	if err != nil {
		return err
	}

	tm.Returns = ReturnMetadata{}
	tm.Returns.Schema = jtm.ReturnsSchema

	return nil
}

// MarshalJSON handles converting TransactionMetadata to JSON since returns is flattened
// in swagger
func (tm *TransactionMetadata) MarshalJSON() ([]byte, error) {
	jtm := jsonTransactionMetadata{tmAlias: (*tmAlias)(tm), ReturnsSchema: tm.Returns.Schema}

	return json.Marshal(&jtm)
}

// ContactMetadata contains contact details about an author of a contract/chaincode
type ContactMetadata struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// LicenseMetadata contains licensing information for contract/chaincode
type LicenseMetadata struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// InfoMetadata contains additional information to clarify use of contract/chaincode
type InfoMetadata struct {
	Description string           `json:"description,omitempty"`
	Title       string           `json:"title,omitempty"`
	Contact     *ContactMetadata `json:"contact,omitempty"`
	License     *LicenseMetadata `json:"license,omitempty"`
	Version     string           `json:"version,omitempty"`
}

// ContractMetadata contains information about what makes up a contract
type ContractMetadata struct {
	Info         *InfoMetadata         `json:"info,omitempty"`
	Name         string                `json:"name"`
	Transactions []TransactionMetadata `json:"transactions"`
	Default      bool                  `json:"default"`
}

// ObjectMetadata description of a component
type ObjectMetadata struct {
	ID                   string                 `json:"$id"`
	Properties           map[string]spec.Schema `json:"properties"`
	Required             []string               `json:"required,omitempty"`
	AdditionalProperties bool                   `json:"additionalProperties"`
}

// ComponentMetadata stores map of schemas of all components
type ComponentMetadata struct {
	Schemas map[string]ObjectMetadata `json:"schemas,omitempty"`
}

// ContractChaincodeMetadata describes a chaincode made using the contract api
type ContractChaincodeMetadata struct {
	Info       *InfoMetadata               `json:"info,omitempty"`
	Contracts  map[string]ContractMetadata `json:"contracts"`
	Components ComponentMetadata           `json:"components"`
}

// Append merge two sets of metadata. Source value will override the original
// values only in fields that are not yet set i.e. when info nil, contracts nil or
// zero length array, components empty.
func (ccm *ContractChaincodeMetadata) Append(source ContractChaincodeMetadata) {
	if ccm.Info == nil {
		ccm.Info = source.Info
	}

	if len(ccm.Contracts) == 0 {
		if ccm.Contracts == nil {
			ccm.Contracts = make(map[string]ContractMetadata)
		}

		for key, value := range source.Contracts {
			ccm.Contracts[key] = value
		}
	}

	if reflect.DeepEqual(ccm.Components, ComponentMetadata{}) {
		ccm.Components = source.Components
	}
}

func compileTransaction(tx TransactionMetadata, contractName string, components ComponentMetadata) (TransactionMetadata, error) {
	for paramIdx, param := range tx.Parameters {
		validator, err := schema.CompileSchema(param.Name, param.Schema, components)
		if err != nil {
			return tx, fmt.Errorf("error compiling schema for %s [%s]. %s schema invalid. %s", contractName, tx.Name, param.Name, err.Error())
		}

		param.Validator = validator
		tx.Parameters[paramIdx] = param
	}

	if tx.Returns.Schema != nil {
		validator, err := schema.CompileSchema("return", tx.Returns.Schema, components)
		if err != nil {
			return tx, fmt.Errorf("error compiling schema for %s [%s]. Return schema invalid. %s", contractName, tx.Name, err.Error())
		}

		tx.Returns.Validator = validator
	}

	return tx, nil
}

// CompileSchemas compiles parameter and return schemas for use by the schema validator. When validating against the
// compiled schema you will need to make the comparison json have a key of the parameter name for parameters or return
// for return values e.g {"param1": "value"}. The compilation process resolves references to components
func (ccm *ContractChaincodeMetadata) CompileSchemas() error {
	for contractName, contract := range ccm.Contracts {
		for txIdx, tx := range contract.Transactions {
			compiledTx, err := compileTransaction(tx, contractName, ccm.Components)
			if err != nil {
				return err
			}

			contract.Transactions[txIdx] = compiledTx
		}
		ccm.Contracts[contractName] = contract
	}

	return nil
}

// ReadMetadataFile return the contents of metadata file as ContractChaincodeMetadata. If no metadata file can be read,
// an error chain containing fs.ErrNotExist is returned.
func ReadMetadataFile() (ContractChaincodeMetadata, error) {
	workingDir, err := osAbs.Getwd()
	if err != nil {
		return ContractChaincodeMetadata{}, err
	}

	metadataPath := filepath.Join(workingDir, MetadataFolder, MetadataFile)
	metadataBytes, err := osAbs.ReadFile(metadataPath)
	if err != nil {
		metadataPath = filepath.Join(workingDir, MetadataFolderSecondary, MetadataFile)
		var err2 error
		metadataBytes, err2 = osAbs.ReadFile(metadataPath)
		if err2 != nil {
			return ContractChaincodeMetadata{}, fmt.Errorf("failed to read metadata from file: %w", errors.Join(err, err2))
		}
	}

	fileMetadata := ContractChaincodeMetadata{}
	fileMetadata.Contracts = make(map[string]ContractMetadata)

	if err := json.Unmarshal(metadataBytes, &fileMetadata); err != nil {
		return ContractChaincodeMetadata{}, err
	}

	return fileMetadata, nil
}

// ValidateAgainstSchema takes a ContractChaincodeMetadata and runs it against the
// schema that defines valid metadata structure. If it does not meet the schema it
// returns an error detailing why
func ValidateAgainstSchema(metadata ContractChaincodeMetadata) error {
	cachedSchema, err := getCompiledSchema()
	if err != nil {
		return err
	}

	return schema.ValidateMetadata(metadata, cachedSchema)
}
