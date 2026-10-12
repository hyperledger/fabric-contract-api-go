// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/go-openapi/spec"
	"github.com/hyperledger/fabric-contract-api-go/v2/internal/utils"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Validator validates a property value against a compiled schema.
type Validator interface {
	Validate(propName string, typ reflect.Type, stringValue string, obj any) error
}

type jsonSchemaValidator struct {
	schema *jsonschema.Schema
}

func (v *jsonSchemaValidator) Validate(propName string, typ reflect.Type, stringValue string, obj any) error {
	if v == nil || v.schema == nil || typ == nil {
		return nil
	}

	toValidate := make(map[string]any)

	if typ == reflect.TypeOf(time.Time{}) {
		toValidate[propName] = stringValue
	} else if typ.Kind() == reflect.Struct || (typ.Kind() == reflect.Pointer && typ.Elem().Kind() == reflect.Struct) {
		toValidate[propName] = json.RawMessage(stringValue)
	} else {
		toValidate[propName] = obj
	}

	raw, err := json.Marshal(toValidate)
	if err != nil {
		return err
	}

	val, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return err
	}

	if err := v.schema.Validate(val); err != nil {
		return fmt.Errorf("value did not match schema:\n%s", utils.ValidateErrorsToString(err))
	}

	return nil
}

// CompileProperties compiles multiple property schemas and components into a Validator.
func CompileProperties(properties map[string]*spec.Schema, components any) (Validator, error) {
	combined := make(map[string]any)
	combined["properties"] = properties

	var compWrapper struct {
		Schemas map[string]json.RawMessage `json:"schemas"`
	}
	if components != nil {
		rawCompBytes, err := json.Marshal(components)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(rawCompBytes, &compWrapper); err != nil {
			return nil, err
		}
		combined["components"] = json.RawMessage(rawCompBytes)
	}

	rawJSON, err := json.Marshal(combined)
	if err != nil {
		return nil, err
	}

	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(rawJSON))
	if err != nil {
		return nil, err
	}

	c := jsonschema.NewCompiler()
	c.AssertFormat()
	base := "http://fabric/"
	resURL := base + "properties.json"

	for name, rawComp := range compWrapper.Schemas {
		compDoc, err := jsonschema.UnmarshalJSON(bytes.NewReader(rawComp))
		if err != nil {
			return nil, err
		}
		if err := c.AddResource(base+name, compDoc); err != nil {
			return nil, err
		}
	}

	if err := c.AddResource(resURL, doc); err != nil {
		return nil, err
	}

	compiled, err := c.Compile(resURL)
	if err != nil {
		return nil, err
	}

	return &jsonSchemaValidator{schema: compiled}, nil
}

// CompileSchema compiles a property schema and components into a Validator.
func CompileSchema(propName string, schema *spec.Schema, components any) (Validator, error) {
	return CompileProperties(map[string]*spec.Schema{propName: schema}, components)
}

func compileRawSchema(resName string, raw []byte) (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	c := jsonschema.NewCompiler()
	c.AssertFormat()
	if err := c.AddResource(resName, doc); err != nil {
		return nil, err
	}
	return c.Compile(resName)
}

// ValidateMetadata validates a metadata object against the provided JSON schema.
func ValidateMetadata(metadata any, schemaJSON []byte) error {
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	schema, err := compileRawSchema("schema.json", schemaJSON)
	if err != nil {
		return err
	}

	instanceDoc, err := jsonschema.UnmarshalJSON(bytes.NewReader(metadataBytes))
	if err != nil {
		return err
	}

	if err := schema.Validate(instanceDoc); err != nil {
		return fmt.Errorf("cannot use metadata. Metadata did not match schema:\n%s", utils.ValidateErrorsToString(err))
	}

	return nil
}
