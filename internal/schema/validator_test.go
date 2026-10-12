// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/require"
)

func TestCompileSchema_Basic(t *testing.T) {
	v, err := CompileSchema("prop", spec.Int64Property(), nil)
	require.NoError(t, err)

	tests := map[string]struct {
		typ         reflect.Type
		stringValue string
		obj         any
		errContains string
	}{
		"valid int": {
			typ:         reflect.TypeOf(int64(42)),
			stringValue: "42",
			obj:         int64(42),
		},
		"invalid int": {
			typ:         reflect.TypeOf("not-a-number"),
			stringValue: "not-a-number",
			obj:         "not-a-number",
			errContains: "value did not match schema",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := v.Validate("prop", tt.typ, tt.stringValue, tt.obj)
			if tt.errContains != "" {
				require.ErrorContains(t, err, tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCompileSchema_Time(t *testing.T) {
	v, err := CompileSchema("prop", spec.DateTimeProperty(), nil)
	require.NoError(t, err)

	now := time.Now().UTC()

	tests := map[string]struct {
		stringValue string
		wantErr     bool
	}{
		"valid date-time": {
			stringValue: now.Format(time.RFC3339),
			wantErr:     false,
		},
		"invalid date-time": {
			stringValue: "invalid-date",
			wantErr:     true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := v.Validate("prop", reflect.TypeOf(time.Time{}), tt.stringValue, now)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCompileSchema_NilValidator(t *testing.T) {
	t.Run("nil validator returns no error", func(t *testing.T) {
		var v *jsonSchemaValidator
		err := v.Validate("prop", reflect.TypeOf(""), "", "")
		require.NoError(t, err)
	})
}

func TestCompileSchema_WithComponents(t *testing.T) {
	components := map[string]any{
		"schemas": map[string]any{
			"Address": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"street": map[string]any{"type": "string"},
				},
				"required": []string{"street"},
			},
		},
	}

	propSchema := spec.RefProperty("http://fabric/Address")
	v, err := CompileSchema("address", propSchema, components)
	require.NoError(t, err)

	tests := map[string]struct {
		value       map[string]any
		stringValue string
		wantErr     bool
	}{
		"valid address": {
			value:       map[string]any{"street": "Main St"},
			stringValue: `{"street":"Main St"}`,
			wantErr:     false,
		},
		"invalid address": {
			value:       map[string]any{"number": 123},
			stringValue: `{"number":123}`,
			wantErr:     true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := v.Validate("address", reflect.TypeOf(tt.value), tt.stringValue, tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type samplePerson struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestCompileSchema_Struct(t *testing.T) {
	personSchema := new(spec.Schema)
	personSchema.Typed("object", "")
	personSchema.AddRequired("name")
	nameProp := spec.StringProperty()
	ageProp := spec.Int64Property()
	personSchema.SetProperty("name", *nameProp)
	personSchema.SetProperty("age", *ageProp)

	v, err := CompileSchema("person", personSchema, nil)
	require.NoError(t, err)

	p := samplePerson{Name: "Alice", Age: 30}

	tests := map[string]struct {
		typ         reflect.Type
		stringValue string
		obj         any
		errContains string
	}{
		"valid struct value": {
			typ:         reflect.TypeOf(p),
			stringValue: `{"name":"Alice","age":30}`,
			obj:         p,
		},
		"valid struct pointer": {
			typ:         reflect.TypeOf(&p),
			stringValue: `{"name":"Alice","age":30}`,
			obj:         &p,
		},
		"invalid struct missing required property": {
			typ:         reflect.TypeOf(p),
			stringValue: `{"age":30}`,
			obj:         p,
			errContains: "missing property 'name'",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := v.Validate("person", tt.typ, tt.stringValue, tt.obj)
			if tt.errContains != "" {
				require.ErrorContains(t, err, tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCompileSchema_NilType(t *testing.T) {
	v, err := CompileSchema("prop", spec.StringProperty(), nil)
	require.NoError(t, err)

	t.Run("nil type returns no error", func(t *testing.T) {
		err := v.Validate("prop", nil, "", nil)
		require.NoError(t, err)
	})
}

func TestCompileRawSchema(t *testing.T) {
	t.Run("valid schema", func(t *testing.T) {
		s, err := CompileRawSchema("test.json", []byte(`{"type": "string"}`))
		require.NoError(t, err)
		require.NotNil(t, s)
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := CompileRawSchema("test.json", []byte(`invalid json`))
		require.Error(t, err)
	})
}

func TestValidateMetadata(t *testing.T) {
	schemaJSON := []byte(`{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"required": ["info"],
		"properties": {
			"info": { "type": "string" }
		}
	}`)

	compiledSchema, err := CompileRawSchema("schema.json", schemaJSON)
	require.NoError(t, err)

	tests := map[string]struct {
		metadata    any
		errContains string
	}{
		"valid metadata": {
			metadata: map[string]any{"info": "test"},
		},
		"invalid metadata": {
			metadata:    map[string]any{"info": 123},
			errContains: "cannot use metadata. Metadata did not match schema",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := ValidateMetadata(tt.metadata, compiledSchema)
			if tt.errContains != "" {
				require.ErrorContains(t, err, tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
