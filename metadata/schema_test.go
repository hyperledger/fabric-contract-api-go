// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package metadata

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/hyperledger/fabric-contract-api-go/internal/types"

	"github.com/stretchr/testify/assert"
)

// ================================
// HELPERS
// ================================

type EmbededType struct {
	Prop0 string
}

type simpleStruct struct {
	Prop1 string
	prop2 string
	prop3 string `metadata:"propname"`
	Prop4 string `json:"jsonname" metadata:",optional"`
	Prop5 string `json:"-"`
	Prop6 string `metadata:"-"`
	Prop7 string `metadata:",optional"`
	Prop8 string `metadata:"prop8, optional"`
	Prop9 string `json:"jsonname2,omitempty"`
}

var simpleStructPropertiesMap = map[string]spec.Schema{
	"Prop1":     *spec.StringProperty(),
	"propname":  *spec.StringProperty(),
	"jsonname":  *spec.StringProperty(),
	"Prop5":     *spec.StringProperty(),
	"Prop7":     *spec.StringProperty(),
	"prop8":     *spec.StringProperty(),
	"jsonname2": *spec.StringProperty(),
}

var simpleStructMetadata = ObjectMetadata{
	ID:                   "simpleStruct",
	Properties:           simpleStructPropertiesMap,
	Required:             []string{"Prop1", "propname", "Prop5", "jsonname2"},
	AdditionalProperties: false,
}

type complexStruct struct {
	EmbededType
	Prop1 string
	Prop2 simpleStruct
	Prop3 *complexStruct `metadata:",optional"`
}

var complexStructPropertiesMap = map[string]spec.Schema{
	"Prop0": *spec.StringProperty(),
	"Prop1": *spec.StringProperty(),
	"Prop2": *spec.RefSchema("simpleStruct"),
	"Prop3": *spec.RefSchema("complexStruct"),
}

var complexStructMetadata = ObjectMetadata{
	ID:                   "complexStruct",
	Properties:           complexStructPropertiesMap,
	Required:             []string{"Prop0", "Prop1", "Prop2"},
	AdditionalProperties: false,
}

type superComplexStruct struct {
	complexStruct
	Prop4 []complexStruct
	Prop5 [2]simpleStruct
	Prop6 map[string]*complexStruct
	Prop7 map[string][]*simpleStruct
}

var superComplexStructPropertiesMap = map[string]spec.Schema{
	"Prop0": *spec.StringProperty(),
	"Prop1": *spec.StringProperty(),
	"Prop2": *spec.RefSchema("simpleStruct"),
	"Prop3": *spec.RefSchema("complexStruct"),
	"Prop4": *spec.ArrayProperty(spec.RefSchema("complexStruct")),
	"Prop5": *spec.ArrayProperty(spec.RefSchema("simpleStruct")),
	"Prop6": *spec.MapProperty(spec.RefSchema("complexStruct")),
	"Prop7": *spec.MapProperty(spec.ArrayProperty(spec.RefSchema("simpleStruct"))),
}

var superComplexStructMetadata = ObjectMetadata{
	ID:                   "superComplexStruct",
	Properties:           superComplexStructPropertiesMap,
	Required:             append(complexStructMetadata.Required, "Prop4", "Prop5", "Prop6", "Prop7"),
	AdditionalProperties: false,
}

type badStruct struct {
	Prop1 complex64
}

var badType = reflect.TypeOf(complex64(1))
var badArrayType = reflect.TypeOf([1]complex64{})
var badSliceType = reflect.TypeOf([]complex64{})
var badMapItemType = reflect.TypeOf(map[string]complex64{})
var badMapKeyType = reflect.TypeOf(map[complex64]string{})

var boolRefType = reflect.TypeOf(true)
var stringRefType = reflect.TypeOf("")
var intRefType = reflect.TypeOf(1)
var int8RefType = reflect.TypeOf(int8(1))
var int16RefType = reflect.TypeOf(int16(1))
var int32RefType = reflect.TypeOf(int32(1))
var int64RefType = reflect.TypeOf(int64(1))
var uintRefType = reflect.TypeOf(uint(1))
var uint8RefType = reflect.TypeOf(uint8(1))
var uint16RefType = reflect.TypeOf(uint16(1))
var uint32RefType = reflect.TypeOf(uint32(1))
var uint64RefType = reflect.TypeOf(uint64(1))
var float32RefType = reflect.TypeOf(float32(1.0))
var float64RefType = reflect.TypeOf(1.0)

func testGetSchema(t *testing.T, typ reflect.Type, expectedSchema *spec.Schema) {
	var schema *spec.Schema
	var err error

	t.Helper()

	schema, err = GetSchema(typ, nil)

	assert.Nil(t, err, fmt.Sprintf("err should be nil for type (%s)", typ.Name()))
	assert.Equal(t, expectedSchema, schema, fmt.Sprintf("should return expected schema for type (%s)", typ.Name()))
}

// ================================
// TESTS
// ================================

func TestBuildArraySchema(t *testing.T) {
	var schema *spec.Schema
	var err error

	zeroArr := [0]int{}
	schema, err = buildArraySchema(reflect.ValueOf(zeroArr), nil, false)
	assert.Equal(t, errors.New("Arrays must have length greater than 0"), err, "should throw error when 0 length array passed")
	assert.Nil(t, schema, "should not have returned a schema for zero array")

	schema, err = buildArraySchema(reflect.ValueOf([1]complex128{}), nil, false)
	_, expectedErr := getSchema(reflect.TypeOf(complex128(1)), nil, false)
	assert.Nil(t, schema, "spec should be nil when GetSchema fails for array")
	assert.Equal(t, expectedErr, err, "should have same error as GetSchema for array")

	schema, err = buildArraySchema(reflect.ValueOf([1]string{}), nil, false)
	expectedLowerSchema, _ := getSchema(reflect.TypeOf(""), nil, false)
	assert.Nil(t, err, "should not error for valid array")
	assert.Equal(t, spec.ArrayProperty(expectedLowerSchema), schema, "should return array of lower schema")
}

func TestBuildSliceSchema(t *testing.T) {
	var schema *spec.Schema
	var err error

	schema, err = buildSliceSchema(reflect.ValueOf([]complex128{}), nil, false)
	_, expectedErr := GetSchema(reflect.TypeOf(complex128(1)), nil)
	assert.Nil(t, schema, "spec should be nil when GetSchema errors for slice")
	assert.Equal(t, expectedErr, err, "should have same error as Getschema for slice")

	schema, err = buildSliceSchema(reflect.ValueOf([]string{}), nil, false)
	expectedLowerSchema, _ := getSchema(reflect.TypeOf(""), nil, false)
	assert.Nil(t, err, "should not error for valid slice")
	assert.Equal(t, spec.ArrayProperty(expectedLowerSchema), schema, "should return spec array of lower schema for slice")
}

func TestBuildMapSchema(t *testing.T) {
	var schema *spec.Schema
	var err error

	schema, err = buildMapSchema(reflect.ValueOf(make(map[string]complex128)), nil, false)
	_, expectedErr := getSchema(reflect.TypeOf(complex128(1)), nil, false)
	assert.Nil(t, schema, "spec should be nil when GetSchema errors for map")
	assert.Equal(t, expectedErr, err, "should have same error as Getschema for map")

	schema, err = buildMapSchema(reflect.ValueOf(make(map[string]string)), nil, false)
	expectedLowerSchema, _ := getSchema(reflect.TypeOf(""), nil, false)
	assert.Nil(t, err, "should not error for valid map")
	assert.Equal(t, spec.MapProperty(expectedLowerSchema), schema, "should return spec map of lower schema")
}

func TestAddComponentIfNotExists(t *testing.T) {
	var err error
	var components *ComponentMetadata
	var ok bool

	someObject := ObjectMetadata{}
	someObject.Properties = make(map[string]spec.Schema)
	someObject.Properties["some property"] = spec.Schema{}

	components = new(ComponentMetadata)
	components.Schemas = make(map[string]ObjectMetadata)
	components.Schemas["simpleStruct"] = someObject

	err = addComponentIfNotExists(reflect.TypeOf(simpleStruct{}), components)
	assert.Nil(t, err, "should return nil for error when component of name already exists")
	assert.Equal(t, len(components.Schemas), 1, "should not have added a new component when one already exists")
	_, ok = components.Schemas["simpleStruct"].Properties["some property"]
	assert.True(t, ok, "should not overwrite existing component")

	err = addComponentIfNotExists(reflect.TypeOf(new(simpleStruct)), components)
	assert.Nil(t, err, "should return nil for error when component of name already exists for pointer")
	assert.Equal(t, len(components.Schemas), 1, "should not have added a new component when one already exists for pointer")
	_, ok = components.Schemas["simpleStruct"].Properties["some property"]
	assert.True(t, ok, "should not overwrite existing component when already exists and pointer passed")

	err = addComponentIfNotExists(reflect.TypeOf(badStruct{}), components)
	_, expectedError := GetSchema(reflect.TypeOf(complex64(1)), components)
	assert.EqualError(t, err, expectedError.Error(), "should use the same error as GetSchema when GetSchema errors")

	components.Schemas = nil
	err = addComponentIfNotExists(reflect.TypeOf(simpleStruct{}), components)
	assert.Nil(t, err, "should not error when adding new component when schemas not initialised")
	assert.Equal(t, components.Schemas["simpleStruct"], simpleStructMetadata, "should set correct metadata for new component when schemas not initialised")

	delete(components.Schemas, "simpleStruct")
	components.Schemas["otherStruct"] = someObject
	err = addComponentIfNotExists(reflect.TypeOf(simpleStruct{}), components)
	assert.Nil(t, err, "should not error when adding new component")
	assert.Equal(t, components.Schemas["simpleStruct"], simpleStructMetadata, "should set correct metadata for new component")
	assert.Equal(t, components.Schemas["otherStruct"], someObject, "should not affect existing components")
}

func TestBuildStructSchema(t *testing.T) {
	var schema *spec.Schema
	var err error

	components := new(ComponentMetadata)
	components.Schemas = make(map[string]ObjectMetadata)

	schema, err = buildStructSchema(reflect.TypeOf(badStruct{}), components, false)
	expectedErr := addComponentIfNotExists(reflect.TypeOf(badStruct{}), components)
	assert.Nil(t, schema, "spec should be nil when buildStructSchema fails from addComponentIfNotExists")
	assert.NotNil(t, err, "error should not be nil")
	assert.Equal(t, expectedErr, err, "should have same error as addComponentIfNotExists")

	schema, err = buildStructSchema(reflect.TypeOf(simpleStruct{}), components, false)
	assert.Nil(t, err, "should not return error when struct is good")
	assert.Equal(t, schema, spec.RefSchema("#/components/schemas/simpleStruct"), "should make schema ref to component")
	_, ok := components.Schemas["simpleStruct"]
	assert.True(t, ok, "should have added component")

	schema, err = buildStructSchema(reflect.TypeOf(simpleStruct{}), components, true)
	assert.Nil(t, err, "should not return error when struct is good")
	assert.Equal(t, schema, spec.RefSchema("simpleStruct"), "should make schema ref to component for nested ref")
	_, ok = components.Schemas["simpleStruct"]
	assert.True(t, ok, "should have added component for nested ref")

	schema, err = buildStructSchema(reflect.TypeOf(new(simpleStruct)), components, false)
	assert.Nil(t, err, "should not return error when pointer to struct is good")
	assert.Equal(t, schema, spec.RefSchema("#/components/schemas/simpleStruct"), "should make schema ref to component")

	_, ok = components.Schemas["simpleStruct"]
	assert.True(t, ok, "should have use already added component")
}

func TestGetSchema(t *testing.T) {
	var schema *spec.Schema
	var err error
	var expectedErr error

	components := new(ComponentMetadata)

	schema, err = GetSchema(badType, components)
	assert.EqualError(t, err, "complex64 was not a valid type", "should return error for invalid type")
	assert.Nil(t, schema, "should return no schema for bad type")

	schema, err = GetSchema(badArrayType, components)
	_, expectedErr = buildArraySchema(reflect.New(badArrayType).Elem(), components, false)
	assert.EqualError(t, err, expectedErr.Error(), "should return error when build array errors")
	assert.Nil(t, schema, "should return no schema when build array errors")

	schema, err = GetSchema(badSliceType, components)
	_, expectedErr = buildSliceSchema(reflect.MakeSlice(badSliceType, 1, 1), components, false)
	assert.EqualError(t, err, expectedErr.Error(), "should return error when build slice errors")
	assert.Nil(t, schema, "should return no schema when build slice errors")

	schema, err = GetSchema(badMapItemType, components)
	_, expectedErr = buildMapSchema(reflect.MakeMap(badMapItemType), components, false)
	assert.EqualError(t, err, expectedErr.Error(), "should return error when build map errors")
	assert.Nil(t, schema, "should return no schema when build map errors")

	schema, err = GetSchema(reflect.TypeOf(badStruct{}), components)
	_, expectedErr = buildStructSchema(reflect.TypeOf(badStruct{}), components, false)
	assert.EqualError(t, err, expectedErr.Error(), "should return error when build struct errors")
	assert.Nil(t, schema, "should return no schema when build struct errors")

	// Test basic types
	testGetSchema(t, stringRefType, types.BasicTypes[reflect.String].GetSchema())
	testGetSchema(t, boolRefType, types.BasicTypes[reflect.Bool].GetSchema())
	testGetSchema(t, intRefType, types.BasicTypes[reflect.Int].GetSchema())
	testGetSchema(t, int8RefType, types.BasicTypes[reflect.Int8].GetSchema())
	testGetSchema(t, int16RefType, types.BasicTypes[reflect.Int16].GetSchema())
	testGetSchema(t, int32RefType, types.BasicTypes[reflect.Int32].GetSchema())
	testGetSchema(t, int64RefType, types.BasicTypes[reflect.Int64].GetSchema())
	testGetSchema(t, uintRefType, types.BasicTypes[reflect.Uint].GetSchema())
	testGetSchema(t, uint8RefType, types.BasicTypes[reflect.Uint8].GetSchema())
	testGetSchema(t, uint16RefType, types.BasicTypes[reflect.Uint16].GetSchema())
	testGetSchema(t, uint32RefType, types.BasicTypes[reflect.Uint32].GetSchema())
	testGetSchema(t, uint64RefType, types.BasicTypes[reflect.Uint64].GetSchema())
	testGetSchema(t, float32RefType, types.BasicTypes[reflect.Float32].GetSchema())
	testGetSchema(t, float64RefType, types.BasicTypes[reflect.Float64].GetSchema())

	// Test advanced types
	testGetSchema(t, types.TimeType, spec.DateTimeProperty())

	// Should return schema for arrays made of each of the valid types
	stringArraySchema := spec.ArrayProperty(types.BasicTypes[reflect.String].GetSchema())
	boolArraySchema := spec.ArrayProperty(types.BasicTypes[reflect.Bool].GetSchema())
	intArraySchema := spec.ArrayProperty(types.BasicTypes[reflect.Int].GetSchema())
	int8ArraySchema := spec.ArrayProperty(types.BasicTypes[reflect.Int8].GetSchema())
	int16ArraySchema := spec.ArrayProperty(types.BasicTypes[reflect.Int16].GetSchema())
	int32ArraySchema := spec.ArrayProperty(types.BasicTypes[reflect.Int32].GetSchema())
	int64ArraySchema := spec.ArrayProperty(types.BasicTypes[reflect.Int64].GetSchema())
	uintArraySchema := spec.ArrayProperty(types.BasicTypes[reflect.Uint].GetSchema())
	uint8ArraySchema := spec.ArrayProperty(types.BasicTypes[reflect.Uint8].GetSchema())
	uint16ArraySchema := spec.ArrayProperty(types.BasicTypes[reflect.Uint16].GetSchema())
	uint32ArraySchema := spec.ArrayProperty(types.BasicTypes[reflect.Uint32].GetSchema())
	uint64ArraySchema := spec.ArrayProperty(types.BasicTypes[reflect.Uint64].GetSchema())
	float32ArraySchema := spec.ArrayProperty(types.BasicTypes[reflect.Float32].GetSchema())
	float64ArraySchema := spec.ArrayProperty(types.BasicTypes[reflect.Float64].GetSchema())

	testGetSchema(t, reflect.TypeOf([1]string{}), stringArraySchema)
	testGetSchema(t, reflect.TypeOf([1]bool{}), boolArraySchema)
	testGetSchema(t, reflect.TypeOf([1]int{}), intArraySchema)
	testGetSchema(t, reflect.TypeOf([1]int8{}), int8ArraySchema)
	testGetSchema(t, reflect.TypeOf([1]int16{}), int16ArraySchema)
	testGetSchema(t, reflect.TypeOf([1]int32{}), int32ArraySchema)
	testGetSchema(t, reflect.TypeOf([1]int64{}), int64ArraySchema)
	testGetSchema(t, reflect.TypeOf([1]uint{}), uintArraySchema)
	testGetSchema(t, reflect.TypeOf([1]uint8{}), uint8ArraySchema)
	testGetSchema(t, reflect.TypeOf([1]uint16{}), uint16ArraySchema)
	testGetSchema(t, reflect.TypeOf([1]uint32{}), uint32ArraySchema)
	testGetSchema(t, reflect.TypeOf([1]uint64{}), uint64ArraySchema)
	testGetSchema(t, reflect.TypeOf([1]float32{}), float32ArraySchema)
	testGetSchema(t, reflect.TypeOf([1]float64{}), float64ArraySchema)
	testGetSchema(t, reflect.TypeOf([1]byte{}), uint8ArraySchema)
	testGetSchema(t, reflect.TypeOf([1]rune{}), int32ArraySchema)

	// Should return schema for multidimensional arrays made of each of the basic types
	testGetSchema(t, reflect.TypeOf([1][1]string{}), spec.ArrayProperty(stringArraySchema))
	testGetSchema(t, reflect.TypeOf([1][1]bool{}), spec.ArrayProperty(boolArraySchema))
	testGetSchema(t, reflect.TypeOf([1][1]int{}), spec.ArrayProperty(intArraySchema))
	testGetSchema(t, reflect.TypeOf([1][1]int8{}), spec.ArrayProperty(int8ArraySchema))
	testGetSchema(t, reflect.TypeOf([1][1]int16{}), spec.ArrayProperty(int16ArraySchema))
	testGetSchema(t, reflect.TypeOf([1][1]int32{}), spec.ArrayProperty(int32ArraySchema))
	testGetSchema(t, reflect.TypeOf([1][1]int64{}), spec.ArrayProperty(int64ArraySchema))
	testGetSchema(t, reflect.TypeOf([1][1]uint{}), spec.ArrayProperty(uintArraySchema))
	testGetSchema(t, reflect.TypeOf([1][1]uint8{}), spec.ArrayProperty(uint8ArraySchema))
	testGetSchema(t, reflect.TypeOf([1][1]uint16{}), spec.ArrayProperty(uint16ArraySchema))
	testGetSchema(t, reflect.TypeOf([1][1]uint32{}), spec.ArrayProperty(uint32ArraySchema))
	testGetSchema(t, reflect.TypeOf([1][1]uint64{}), spec.ArrayProperty(uint64ArraySchema))
	testGetSchema(t, reflect.TypeOf([1][1]float32{}), spec.ArrayProperty(float32ArraySchema))
	testGetSchema(t, reflect.TypeOf([1][1]float64{}), spec.ArrayProperty(float64ArraySchema))
	testGetSchema(t, reflect.TypeOf([1][1]byte{}), spec.ArrayProperty(uint8ArraySchema))
	testGetSchema(t, reflect.TypeOf([1][1]rune{}), spec.ArrayProperty(int32ArraySchema))

	// Should build schema for a big multidimensional array
	testGetSchema(t, reflect.TypeOf([1][2][3][4][5][6][7][8]string{}), spec.ArrayProperty(spec.ArrayProperty(spec.ArrayProperty(spec.ArrayProperty(spec.ArrayProperty(spec.ArrayProperty(spec.ArrayProperty(stringArraySchema))))))))

	// Should return error when array is not one of the valid types
	badSlice := []complex128{}
	schema, err = GetSchema(reflect.TypeOf(badSlice), nil)

	assert.EqualError(t, err, "complex128 was not a valid type", "should throw error when invalid type passed")
	assert.Nil(t, schema, "should not have returned a schema for an array of bad type")

	// Should return an error when array is passed with sub array with a length of zero
	zeroSubArrInSlice := [][0]int{}
	schema, err = GetSchema(reflect.TypeOf(zeroSubArrInSlice), nil)

	assert.Equal(t, errors.New("Arrays must have length greater than 0"), err, "should throw error when 0 length array passed")
	assert.Nil(t, schema, "should not have returned a schema for zero array")

	// should build schema for slices of original types
	testGetSchema(t, reflect.TypeOf([]string{""}), stringArraySchema)
	testGetSchema(t, reflect.TypeOf([]bool{true}), boolArraySchema)
	testGetSchema(t, reflect.TypeOf([]int{1}), intArraySchema)
	testGetSchema(t, reflect.TypeOf([]int8{1}), int8ArraySchema)
	testGetSchema(t, reflect.TypeOf([]int16{1}), int16ArraySchema)
	testGetSchema(t, reflect.TypeOf([]int32{1}), int32ArraySchema)
	testGetSchema(t, reflect.TypeOf([]int64{1}), int64ArraySchema)
	testGetSchema(t, reflect.TypeOf([]uint{1}), uintArraySchema)
	testGetSchema(t, reflect.TypeOf([]uint8{1}), uint8ArraySchema)
	testGetSchema(t, reflect.TypeOf([]uint16{1}), uint16ArraySchema)
	testGetSchema(t, reflect.TypeOf([]uint32{1}), uint32ArraySchema)
	testGetSchema(t, reflect.TypeOf([]uint64{1}), uint64ArraySchema)
	testGetSchema(t, reflect.TypeOf([]float32{1}), float32ArraySchema)
	testGetSchema(t, reflect.TypeOf([]float64{1}), float64ArraySchema)
	testGetSchema(t, reflect.TypeOf([]byte{1}), uint8ArraySchema)
	testGetSchema(t, reflect.TypeOf([]rune{1}), int32ArraySchema)

	// Should return schema for multidimensional slices made of each of the basic types
	testGetSchema(t, reflect.TypeOf([][]bool{{}}), spec.ArrayProperty(boolArraySchema))
	testGetSchema(t, reflect.TypeOf([][]int{{}}), spec.ArrayProperty(intArraySchema))
	testGetSchema(t, reflect.TypeOf([][]int8{{}}), spec.ArrayProperty(int8ArraySchema))
	testGetSchema(t, reflect.TypeOf([][]int16{{}}), spec.ArrayProperty(int16ArraySchema))
	testGetSchema(t, reflect.TypeOf([][]int32{{}}), spec.ArrayProperty(int32ArraySchema))
	testGetSchema(t, reflect.TypeOf([][]int64{{}}), spec.ArrayProperty(int64ArraySchema))
	testGetSchema(t, reflect.TypeOf([][]uint{{}}), spec.ArrayProperty(uintArraySchema))
	testGetSchema(t, reflect.TypeOf([][]uint8{{}}), spec.ArrayProperty(uint8ArraySchema))
	testGetSchema(t, reflect.TypeOf([][]uint16{{}}), spec.ArrayProperty(uint16ArraySchema))
	testGetSchema(t, reflect.TypeOf([][]uint32{{}}), spec.ArrayProperty(uint32ArraySchema))
	testGetSchema(t, reflect.TypeOf([][]uint64{{}}), spec.ArrayProperty(uint64ArraySchema))
	testGetSchema(t, reflect.TypeOf([][]float32{{}}), spec.ArrayProperty(float32ArraySchema))
	testGetSchema(t, reflect.TypeOf([][]float64{{}}), spec.ArrayProperty(float64ArraySchema))
	testGetSchema(t, reflect.TypeOf([][]byte{{}}), spec.ArrayProperty(uint8ArraySchema))
	testGetSchema(t, reflect.TypeOf([][]rune{{}}), spec.ArrayProperty(int32ArraySchema))

	// Should handle an array of slice
	testGetSchema(t, reflect.TypeOf([1][]string{}), spec.ArrayProperty(stringArraySchema))

	// Should handle a slice of array
	testGetSchema(t, reflect.TypeOf([][1]string{{}}), spec.ArrayProperty(stringArraySchema))

	// Should handle a map
	testGetSchema(t, reflect.TypeOf(map[string]int{}), spec.MapProperty(types.BasicTypes[reflect.Int].GetSchema()))

	// Should handle a of map map
	testGetSchema(t, reflect.TypeOf(map[string]map[string]int{}), spec.MapProperty(spec.MapProperty(types.BasicTypes[reflect.Int].GetSchema())))

	// Should return error when multidimensional array/slice/array is bad
	badMixedArr := [1][][0]string{}
	schema, err = GetSchema(reflect.TypeOf(badMixedArr), nil)

	assert.EqualError(t, err, "Arrays must have length greater than 0", "should throw error when 0 length array passed")
	assert.Nil(t, schema, "schema should be nil when sub array bad type")

	// Should handle a valid struct and add to components
	components = new(ComponentMetadata)
	components.Schemas = make(map[string]ObjectMetadata)

	schema, err = GetSchema(reflect.TypeOf(simpleStruct{}), components)

	assert.Nil(t, err, "should return nil when valid object")
	assert.Equal(t, len(components.Schemas), 1, "should have added a new component")
	assert.Equal(t, components.Schemas["simpleStruct"], simpleStructMetadata, "should have added correct metadata to components")
	assert.Equal(t, schema, spec.RefSchema("#/components/schemas/simpleStruct"))

	// should handle pointer to struct
	components = new(ComponentMetadata)
	components.Schemas = make(map[string]ObjectMetadata)

	schema, err = GetSchema(reflect.TypeOf(new(simpleStruct)), components)

	assert.Nil(t, err, "should return nil when valid object")
	assert.Equal(t, len(components.Schemas), 1, "should have added a new component")
	assert.Equal(t, components.Schemas["simpleStruct"], simpleStructMetadata, "should have added correct metadata to components")
	assert.Equal(t, schema, spec.RefSchema("#/components/schemas/simpleStruct"))

	// Should handle an array of structs
	components = new(ComponentMetadata)
	components.Schemas = make(map[string]ObjectMetadata)

	schema, err = GetSchema(reflect.TypeOf([1]simpleStruct{}), components)

	assert.Nil(t, err, "should return nil when valid object")
	assert.Equal(t, len(components.Schemas), 1, "should have added a new component")
	assert.Equal(t, components.Schemas["simpleStruct"], simpleStructMetadata, "should have added correct metadata to components")
	assert.Equal(t, schema, spec.ArrayProperty(spec.RefSchema("#/components/schemas/simpleStruct")))

	// Should handle a slice of structs
	components = new(ComponentMetadata)
	components.Schemas = make(map[string]ObjectMetadata)

	schema, err = GetSchema(reflect.TypeOf([]simpleStruct{}), components)

	assert.Nil(t, err, "should return nil when valid object")
	assert.Equal(t, len(components.Schemas), 1, "should have added a new component")
	assert.Equal(t, components.Schemas["simpleStruct"], simpleStructMetadata, "should have added correct metadata to components")
	assert.Equal(t, schema, spec.ArrayProperty(spec.RefSchema("#/components/schemas/simpleStruct")))

	// Should handle a valid struct with struct property and add to components
	components = new(ComponentMetadata)
	components.Schemas = make(map[string]ObjectMetadata)

	schema, err = GetSchema(reflect.TypeOf(new(complexStruct)), components)

	assert.Nil(t, err, "should return nil when valid object")
	assert.Equal(t, len(components.Schemas), 2, "should have added two new components")
	assert.Equal(t, components.Schemas["simpleStruct"], simpleStructMetadata, "should have added correct metadata to components for sub struct")
	assert.Equal(t, components.Schemas["complexStruct"], complexStructMetadata, "should have added correct metadata to components for main struct")
	assert.Equal(t, schema, spec.RefSchema("#/components/schemas/complexStruct"))

	// Should handle a valid struct with struct properties of array, slice and map types
	components = new(ComponentMetadata)
	components.Schemas = make(map[string]ObjectMetadata)

	schema, err = GetSchema(reflect.TypeOf(new(superComplexStruct)), components)

	assert.Nil(t, err, "should return nil when valid object")
	assert.Equal(t, len(components.Schemas), 3, "should have added two new components")
	assert.Equal(t, components.Schemas["simpleStruct"], simpleStructMetadata, "should have added correct metadata to components for sub struct")
	assert.Equal(t, components.Schemas["complexStruct"], complexStructMetadata, "should have added correct metadata to components for sub struct")
	assert.Equal(t, components.Schemas["superComplexStruct"], superComplexStructMetadata, "should have added correct metadata to components for main struct")
	assert.Equal(t, schema, spec.RefSchema("#/components/schemas/superComplexStruct"))

	// Should return an error for a bad struct
	components = new(ComponentMetadata)
	components.Schemas = make(map[string]ObjectMetadata)

	schema, err = GetSchema(reflect.TypeOf(new(badStruct)), components)

	assert.Nil(t, schema, "should not give back a schema when struct is bad")
	assert.EqualError(t, err, "complex64 was not a valid type", "should return err when invalid object")
	assert.Equal(t, len(components.Schemas), 0, "should not have added new component")
}
