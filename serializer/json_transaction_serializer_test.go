// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package serializer

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/go-openapi/spec"
	"github.com/hyperledger/fabric-contract-api-go/internal/types"
	"github.com/hyperledger/fabric-contract-api-go/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

// ================================
// HELPERS
// ================================

type simpleStruct struct {
	Prop1 string `json:"prop1"`
	prop2 string
}

type UsefulInterface interface{}

type usefulStruct struct {
	ptr      *string
	iface    UsefulInterface
	mp       map[string]string
	slice    []string
	channel  chan string
	basic    string
	array    [1]string
	strct    simpleStruct
	strctPtr *simpleStruct
}

func (us usefulStruct) DoNothing() string {
	return "nothing"
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

func testConvertArgsBasicType(t *testing.T, expected interface{}, str string) {
	t.Helper()

	var expectedValue reflect.Value
	var actualValue reflect.Value
	var actualErr error

	typ := reflect.TypeOf(expected)

	expectedValue, _ = types.BasicTypes[typ.Kind()].Convert(str)
	actualValue, actualErr = convertArg(typ, str)
	assert.Nil(t, actualErr, fmt.Sprintf("should not return an error on good convert (%s)", typ.Name()))
	assert.Equal(t, expectedValue.Interface(), actualValue.Interface(), fmt.Sprintf("should return same value as convert for good convert (%s)", typ.Name()))
}

func testConvertArgsComplexType(t *testing.T, expected interface{}, str string) {
	t.Helper()

	var expectedValue reflect.Value
	var actualValue reflect.Value
	var actualErr error

	typ := reflect.TypeOf(expected)

	expectedValue, _ = createArraySliceMapOrStruct(str, typ)
	actualValue, actualErr = convertArg(typ, str)
	assert.Nil(t, actualErr, fmt.Sprintf("should not return an error on good complex convert (%s)", typ.Name()))
	assert.Equal(t, expectedValue.Interface(), actualValue.Interface(), fmt.Sprintf("should return same value as convert for good complex convert (%s)", typ.Name()))
}

// ================================
// TESTS
// ================================

func TestIsNillableType(t *testing.T) {
	usefulStruct := usefulStruct{}
	usefulStructType := reflect.TypeOf(usefulStruct)

	assert.True(t, isNillableType(usefulStructType.Field(0).Type.Kind()), "should return true for pointers")

	assert.True(t, isNillableType(usefulStructType.Field(1).Type.Kind()), "should return true for interfaces")

	assert.True(t, isNillableType(usefulStructType.Field(2).Type.Kind()), "should return true for maps")

	assert.True(t, isNillableType(usefulStructType.Field(3).Type.Kind()), "should return true for slices")

	assert.True(t, isNillableType(usefulStructType.Field(4).Type.Kind()), "should return true for channels")

	assert.True(t, isNillableType(usefulStructType.Method(0).Type.Kind()), "should return true for func")

	assert.False(t, isNillableType(usefulStructType.Field(5).Type.Kind()), "should return false for something that isnt the above")
}

func TestIsMarshallingType(t *testing.T) {
	usefulStruct := usefulStruct{}
	usefulStructType := reflect.TypeOf(usefulStruct)

	assert.True(t, isMarshallingType(usefulStructType.Field(6).Type), "should return true for arrays")

	assert.True(t, isMarshallingType(usefulStructType.Field(3).Type), "should return true for slices")

	assert.True(t, isMarshallingType(usefulStructType.Field(2).Type), "should return true for maps")

	assert.True(t, isMarshallingType(usefulStructType.Field(7).Type), "should return true for structs")

	assert.True(t, isMarshallingType(usefulStructType.Field(8).Type), "should return true for pointer of marshalling type")

	assert.False(t, isMarshallingType(usefulStructType.Field(5).Type), "should return false for something that isnt the above")

	assert.False(t, isMarshallingType(usefulStructType.Field(0).Type), "should return false for pointer to non marshalling type")
}

func TestCreateArraySliceMapOrStruct(t *testing.T) {
	var val reflect.Value
	var err error

	arrType := reflect.TypeOf([1]string{})
	val, err = createArraySliceMapOrStruct("bad json", arrType)
	assert.EqualError(t, err, fmt.Sprintf("Value %s was not passed in expected format %s", "bad json", arrType.String()), "should error when JSON marshall fails")
	assert.Equal(t, reflect.Value{}, val, "should return an empty value when error found")

	val, err = createArraySliceMapOrStruct("[\"array\"]", arrType)
	assert.Nil(t, err, "should not error for valid array JSON")
	assert.Equal(t, [1]string{"array"}, val.Interface().([1]string), "should produce populated array")

	val, err = createArraySliceMapOrStruct("[\"slice\", \"slice\", \"baby\"]", reflect.TypeOf([]string{}))
	assert.Nil(t, err, "should not error for valid slice JSON")
	assert.Equal(t, []string{"slice", "slice", "baby"}, val.Interface().([]string), "should produce populated slice")

	val, err = createArraySliceMapOrStruct("{\"Prop1\": \"value\"}", reflect.TypeOf(simpleStruct{}))
	assert.Nil(t, err, "should not error for valid struct json")
	assert.Equal(t, simpleStruct{"value", ""}, val.Interface().(simpleStruct), "should produce populated struct")

	val, err = createArraySliceMapOrStruct("{\"key\": 1}", reflect.TypeOf(make(map[string]int)))
	assert.Nil(t, err, "should not error for valid map JSON")
	assert.Equal(t, map[string]int{"key": 1}, val.Interface().(map[string]int), "should produce populated map")
}

func TestConvertArg(t *testing.T) {
	var expectedErr error
	var actualValue reflect.Value
	var actualErr error

	_, expectedErr = types.BasicTypes[reflect.Int].Convert("NaN")
	actualValue, actualErr = convertArg(reflect.TypeOf(1), "NaN")
	assert.Equal(t, reflect.Value{}, actualValue, "should not return a value when basic type conversion fails")
	assert.EqualError(t, actualErr, fmt.Sprintf("Conversion error. %s", expectedErr.Error()), "should error on basic type conversion error using message")

	_, expectedErr = createArraySliceMapOrStruct("Not an array", reflect.TypeOf([1]string{}))
	actualValue, actualErr = convertArg(reflect.TypeOf([1]string{}), "Not an array")
	assert.Equal(t, reflect.Value{}, actualValue, "should not return a value when complex type conversion fails")
	assert.EqualError(t, actualErr, fmt.Sprintf("Conversion error. %s", expectedErr.Error()), "should error on complex type conversion error using message")

	// should handle basic types
	testConvertArgsBasicType(t, "some string", "some string")
	testConvertArgsBasicType(t, 1, "1")
	testConvertArgsBasicType(t, int8(1), "1")
	testConvertArgsBasicType(t, int16(1), "1")
	testConvertArgsBasicType(t, int32(1), "1")
	testConvertArgsBasicType(t, int64(1), "1")
	testConvertArgsBasicType(t, uint(1), "1")
	testConvertArgsBasicType(t, uint8(1), "1")
	testConvertArgsBasicType(t, uint16(1), "1")
	testConvertArgsBasicType(t, uint32(1), "1")
	testConvertArgsBasicType(t, uint64(1), "1")
	testConvertArgsBasicType(t, true, "true")

	// should handle array, slice, map and struct
	testConvertArgsComplexType(t, [1]int{}, "[1,2,3]")
	testConvertArgsComplexType(t, []string{}, "[\"a\",\"string\",\"array\"]")
	testConvertArgsComplexType(t, make(map[string]bool), "{\"a\": true, \"b\": false}")
	testConvertArgsComplexType(t, simpleStruct{}, "{\"Prop1\": \"hello\"}")
	testConvertArgsComplexType(t, &simpleStruct{}, "{\"Prop1\": \"hello\"}")

	// should handle time
	actualValue, actualErr = convertArg(types.TimeType, "2002-10-02T15:00:00Z")
	assert.Nil(t, actualErr, "should not return error for RFC3339 type")
	expectedTime, _ := time.Parse(time.RFC3339, "2002-10-02T15:00:00Z")
	assert.Equal(t, expectedTime, actualValue.Interface(), "should create error using string for error type")
}

func TestValidateAgainstSchema(t *testing.T) {
	var comparisonSchema *gojsonschema.Schema
	var err error

	components := new(metadata.ComponentMetadata)

	comparisonSchema = createGoJSONSchemaSchema("prop", types.BasicTypes[reflect.Uint].GetSchema(), components)
	err = validateAgainstSchema("prop", reflect.TypeOf(-1), "-1", -1, comparisonSchema)
	assert.Contains(t, err.Error(), "Value did not match schema", "should error when data doesnt match schema")

	comparisonSchema = createGoJSONSchemaSchema("prop", spec.DateTimeProperty(), nil)
	timeObj, _ := time.Parse(time.RFC3339, "2002-10-02T15:00:00Z")
	err = validateAgainstSchema("prop", types.TimeType, "2002/10/02 15:00:00", timeObj, comparisonSchema)
	assert.Contains(t, err.Error(), "Value did not match schema", "should error for invalid time")

	comparisonSchema = createGoJSONSchemaSchema("prop", types.BasicTypes[reflect.Uint].GetSchema(), components)
	err = validateAgainstSchema("prop", reflect.TypeOf(10), "10", 10, comparisonSchema)
	assert.Nil(t, err, "should not error when matches schema")

	expectedStruct := new(simpleStruct)
	expectedStruct.Prop1 = "hello"
	bytes, _ := json.Marshal(expectedStruct)
	schema, _ := metadata.GetSchema(reflect.TypeOf(expectedStruct), components)
	comparisonSchema = createGoJSONSchemaSchema("prop", schema, components)
	err = validateAgainstSchema("prop", reflect.TypeOf(expectedStruct), string(bytes), expectedStruct, comparisonSchema)
	assert.Nil(t, err, "should handle struct")
}

func TestFromString(t *testing.T) {
	var err error
	var value reflect.Value
	var expectedErr error
	var schema *spec.Schema
	var compiledSchema *gojsonschema.Schema
	var paramMetadata metadata.ParameterMetadata

	serializer := new(JSONSerializer)

	value, err = serializer.FromString("some string", reflect.TypeOf(1), nil, nil)
	_, expectedErr = convertArg(reflect.TypeOf(1), "some string")
	assert.EqualError(t, err, expectedErr.Error(), "should error when convertArg errors")
	assert.Equal(t, reflect.Value{}, value, "should return an empty reflect value when it errors due to convertArg")

	float := float64(2)
	schema = spec.Int64Property()
	schema.Minimum = &float
	compiledSchema = createGoJSONSchemaSchema("param1", schema, nil)
	paramMetadata = metadata.ParameterMetadata{Name: "param1", Schema: schema, CompiledSchema: compiledSchema}
	value, err = serializer.FromString("1", reflect.TypeOf(1), &paramMetadata, nil)
	expectedErr = validateAgainstSchema("param1", reflect.TypeOf(1), "1", 1, compiledSchema)
	assert.EqualError(t, err, expectedErr.Error(), "should error when validateAgainstSchema errors")
	assert.Equal(t, reflect.Value{}, value, "should return an empty reflect value when it errors due to validateAgainstSchema")

	value, err = serializer.FromString("1234", reflect.TypeOf(1), nil, nil)
	assert.Nil(t, err, "should not error when convert args passes and no schema")
	assert.Equal(t, reflect.ValueOf(1234).Interface(), value.Interface(), "should reflect value for converted arg")

	expectedStruct := new(simpleStruct)
	expectedStruct.Prop1 = "hello"
	components := new(metadata.ComponentMetadata)
	schema, _ = metadata.GetSchema(reflect.TypeOf(expectedStruct), components)
	compiledSchema = createGoJSONSchemaSchema("param1", schema, components)
	paramMetadata = metadata.ParameterMetadata{Name: "param1", Schema: schema, CompiledSchema: compiledSchema}
	value, err = serializer.FromString("{\"prop1\":\"hello\"}", reflect.TypeOf(expectedStruct), &paramMetadata, components)
	assert.Nil(t, err, "should not error when convert args passes and schema passes")
	assert.Equal(t, reflect.ValueOf(expectedStruct).Interface(), value.Interface(), "should reflect value for converted arg when arg and schema passes")
}

func TestToString(t *testing.T) {
	var err error
	var value string
	var expectedErr error
	var schema *spec.Schema
	var compiledSchema *gojsonschema.Schema
	var returnMetadata metadata.ReturnMetadata

	serializer := new(JSONSerializer)

	var nilResult *simpleStruct
	value, err = serializer.ToString(reflect.ValueOf(nilResult), reflect.TypeOf(new(simpleStruct)), nil, nil)
	assert.Nil(t, err, "should not error when receives nil")
	assert.Equal(t, "", value, "should return blank string for nil value")

	result := new(simpleStruct)
	result.Prop1 = "property 1"
	value, err = serializer.ToString(reflect.ValueOf(result), reflect.TypeOf(result), nil, nil)
	assert.Nil(t, err, "should not error when receives non nil nillable type")
	assert.Equal(t, "{\"prop1\":\"property 1\"}", value, "should return JSON formatted value for marshallable type")

	value, err = serializer.ToString(reflect.ValueOf(1), reflect.TypeOf(1), nil, nil)
	assert.Nil(t, err, "should not error when receives non nillable and marshalling type")
	assert.Equal(t, "1", value, "should return sprint version of value when not marshalling type")

	testTime, _ := time.Parse(time.RFC3339, "2002-10-02T15:00:00Z")
	value, err = serializer.ToString(reflect.ValueOf(testTime), types.TimeType, nil, nil)
	assert.Nil(t, err, "should not error for time")
	assert.Equal(t, "2002-10-02T15:00:00Z", value, "should return string version of time in RFC3339 format")

	float := float64(2)
	schema = spec.Int64Property()
	schema.Minimum = &float
	compiledSchema = createGoJSONSchemaSchema("return", schema, nil)
	returnMetadata = metadata.ReturnMetadata{Schema: schema, CompiledSchema: compiledSchema}
	value, err = serializer.ToString(reflect.ValueOf(1), reflect.TypeOf(1), &returnMetadata, nil)
	expectedErr = validateAgainstSchema("return", reflect.TypeOf(1), "1", 1, compiledSchema)
	assert.EqualError(t, err, expectedErr.Error(), "should error when validateAgainstSchema errors")
	assert.Equal(t, "", value, "should return an empty string value when it errors due to validateAgainstSchema")

	expectedStruct := new(simpleStruct)
	expectedStruct.Prop1 = "hello"
	components := new(metadata.ComponentMetadata)
	schema, _ = metadata.GetSchema(reflect.TypeOf(expectedStruct), components)
	compiledSchema = createGoJSONSchemaSchema("return", schema, components)
	returnMetadata = metadata.ReturnMetadata{Schema: schema, CompiledSchema: compiledSchema}
	value, err = serializer.ToString(reflect.ValueOf(expectedStruct), reflect.TypeOf(expectedStruct), &returnMetadata, components)
	assert.Nil(t, err, "should not error when making a string passes and schema passes")
	assert.Equal(t, "{\"prop1\":\"hello\"}", value, "should return string value when schema passes")
}
