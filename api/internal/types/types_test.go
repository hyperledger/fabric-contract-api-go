// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
)

// ================================
// HELPERS
// ================================
const convertError = "Cannot convert passed value %s to int"

// ================================
// TESTS
// ================================

func TestMain(m *testing.M) {
	rc := m.Run()

	if rc == 0 && testing.CoverMode() != "" {
		c := testing.Coverage()

		if c < 1 {
			fmt.Println("Tests passed but coverage failed at", c)
			rc = -1
		}
	}

	os.Exit(rc)
}

func TestStringType(t *testing.T) {
	var stringTypeVar = new(stringType)

	// Test GetSchema
	assert.Equal(t, spec.StringProperty(), stringTypeVar.GetSchema(), "should return open api string spec")

	// Test Convert
	val, err := stringTypeVar.Convert("some string")
	assert.Nil(t, err, "should not return error for valid string value")
	assert.Equal(t, "some string", val.Interface().(string), "should have returned the same string")
}

func TestBoolType(t *testing.T) {
	var boolTypeVar = new(boolType)

	// Test GetSchema
	assert.Equal(t, spec.BooleanProperty(), boolTypeVar.GetSchema(), "should return open api bool spec")

	// Test Convert
	var val reflect.Value
	var err error

	val, err = boolTypeVar.Convert("true")
	assert.Nil(t, err, "should not return error for valid bool (true) value")
	assert.True(t, val.Interface().(bool), "should have returned the boolean true")

	val, err = boolTypeVar.Convert("false")
	assert.Nil(t, err, "should not return error for valid bool (false) value")
	assert.False(t, val.Interface().(bool), "should have returned the boolean false")

	val, err = boolTypeVar.Convert("")
	assert.Nil(t, err, "should not return error for valid bool (blank) value")
	assert.False(t, val.Interface().(bool), "should have returned the boolean false for blank value")

	val, err = boolTypeVar.Convert("non bool")
	assert.Error(t, fmt.Errorf(convertError, "non bool"), "should return error for invalid bool value")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for non bool")
}

func TestIntType(t *testing.T) {
	var intTypeVar = new(intType)

	// Test GetSchema
	assert.Equal(t, spec.Int64Property(), intTypeVar.GetSchema(), "should return open api int64 spec")

	// Test Convert
	var val reflect.Value
	var err error

	val, err = intTypeVar.Convert("123")
	assert.Nil(t, err, "should not return error for valid int (123) value")
	assert.Equal(t, 123, val.Interface().(int), "should have returned the int value 123")

	val, err = intTypeVar.Convert("")
	assert.Nil(t, err, "should not return error for valid int (blank) value")
	assert.Equal(t, 0, val.Interface().(int), "should have returned the deafult int value")

	val, err = intTypeVar.Convert("not a number")
	assert.Error(t, err, fmt.Errorf(convertError, "not a number"), "should return error for invalid int value")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid int")
}

func TestInt8Type(t *testing.T) {
	var int8TypeVar = new(int8Type)

	// Test GetSchema
	assert.Equal(t, spec.Int8Property(), int8TypeVar.GetSchema(), "should return open api int8 spec")

	// Test Convert
	var val reflect.Value
	var err error

	val, err = int8TypeVar.Convert("123")
	assert.Nil(t, err, "should not return error for valid int8 (123) value")
	assert.Equal(t, int8(123), val.Interface().(int8), "should have returned the int8 value 123")

	val, err = int8TypeVar.Convert("")
	assert.Nil(t, err, "should not return error for valid int8 (blank) value")
	assert.Equal(t, int8(0), val.Interface().(int8), "should have returned the deafult int8 value")

	val, err = int8TypeVar.Convert("not a number")
	assert.Error(t, err, fmt.Errorf(convertError, "not a number"), "should return error for invalid int8 value (NaN)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid int8 (NaN)")

	tooBig := strconv.Itoa(math.MaxInt8 + 1)
	val, err = int8TypeVar.Convert(tooBig)
	assert.Error(t, err, fmt.Errorf(convertError, tooBig), "should return error for invalid int8 value (too large)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid int8 (too large)")
}

func TestInt16Type(t *testing.T) {
	var int16TypeVar = new(int16Type)

	// Test GetSchema
	assert.Equal(t, spec.Int16Property(), int16TypeVar.GetSchema(), "should return open api int16 spec")

	// Test Convert
	var val reflect.Value
	var err error

	val, err = int16TypeVar.Convert("123")
	assert.Nil(t, err, "should not return error for valid int16 (123) value")
	assert.Equal(t, int16(123), val.Interface().(int16), "should have returned the int16 value 123")

	val, err = int16TypeVar.Convert("")
	assert.Nil(t, err, "should not return error for valid int16 (blank) value")
	assert.Equal(t, int16(0), val.Interface().(int16), "should have returned the deafult int16 value")

	val, err = int16TypeVar.Convert("not a number")
	assert.Error(t, err, fmt.Errorf(convertError, "not a number"), "should return error for invalid int16 value (NaN)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid int16 (NaN)")

	tooBig := strconv.Itoa(math.MaxInt16 + 1)
	val, err = int16TypeVar.Convert(tooBig)
	assert.Error(t, err, fmt.Errorf(convertError, tooBig), "should return error for invalid int16 value (too large)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid int16 (too large)")
}

func TestInt32Type(t *testing.T) {
	var int32TypeVar = new(int32Type)

	// Test GetSchema
	assert.Equal(t, spec.Int32Property(), int32TypeVar.GetSchema(), "should return open api int32 spec")

	// Test Convert
	var val reflect.Value
	var err error

	val, err = int32TypeVar.Convert("123")
	assert.Nil(t, err, "should not return error for valid int32 (123) value")
	assert.Equal(t, int32(123), val.Interface().(int32), "should have returned the int32 value 123")

	val, err = int32TypeVar.Convert("")
	assert.Nil(t, err, "should not return error for valid int32 (blank) value")
	assert.Equal(t, int32(0), val.Interface().(int32), "should have returned the deafult int32 value")

	val, err = int32TypeVar.Convert("not a number")
	assert.Error(t, err, fmt.Errorf(convertError, "not a number"), "should return error for invalid int32 value (NaN)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid int32 (NaN)")

	tooBig := strconv.Itoa(math.MaxInt32 + 1)
	val, err = int32TypeVar.Convert(tooBig)
	assert.Error(t, err, fmt.Errorf(convertError, tooBig), "should return error for invalid int32 value (too large)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid int32 (too large)")
}

func TestInt64Type(t *testing.T) {
	var int64TypeVar = new(int64Type)

	// Test GetSchema
	assert.Equal(t, spec.Int64Property(), int64TypeVar.GetSchema(), "should return open api int64 spec")

	// Test Convert
	var val reflect.Value
	var err error

	val, err = int64TypeVar.Convert("123")
	assert.Nil(t, err, "should not return error for valid int64 (123) value")
	assert.Equal(t, int64(123), val.Interface().(int64), "should have returned the int64 value 123")

	val, err = int64TypeVar.Convert("")
	assert.Nil(t, err, "should not return error for valid int64 (blank) value")
	assert.Equal(t, int64(0), val.Interface().(int64), "should have returned the deafult int64 value")

	val, err = int64TypeVar.Convert("not a number")
	assert.Error(t, err, fmt.Errorf(convertError, "not a number"), "should return error for invalid int64 value (NaN)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid int64 (NaN)")
}

func TestUintType(t *testing.T) {
	var uintTypeVar = new(uintType)

	// Test GetSchema
	expectedSchema := spec.Float64Property()
	multOf := float64(1)
	expectedSchema.MultipleOf = &multOf
	minimum := float64(0)
	expectedSchema.Minimum = &minimum
	maximum := float64(math.MaxUint64)
	expectedSchema.Maximum = &maximum
	actualSchema := uintTypeVar.GetSchema()

	assert.Equal(t, expectedSchema, actualSchema, "should return valid open api format that fits uints")

	// Test Convert
	var val reflect.Value
	var err error

	val, err = uintTypeVar.Convert("123")
	assert.Nil(t, err, "should not return error for valid uint (123) value")
	assert.Equal(t, uint(123), val.Interface().(uint), "should have returned the uint value 123")

	val, err = uintTypeVar.Convert("")
	assert.Nil(t, err, "should not return error for valid uint (blank) value")
	assert.Equal(t, uint(0), val.Interface().(uint), "should have returned the deafult uint value")

	val, err = uintTypeVar.Convert("not a number")
	assert.Error(t, err, fmt.Errorf(convertError, "not a number"), "should return error for invalid uint value (NaN)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid uint (NaN)")

	val, err = uintTypeVar.Convert("-1")
	assert.Error(t, err, fmt.Errorf(convertError, "-1"), "should return error for invalid uint value (-1)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid uint (-1)")
}

func TestUint8Type(t *testing.T) {
	var uint8TypeVar = new(uint8Type)

	// Test GetSchema
	expectedSchema := spec.Int32Property()
	minimum := float64(0)
	expectedSchema.Minimum = &minimum

	maximum := float64(math.MaxUint8)
	expectedSchema.Maximum = &maximum
	actualSchema := uint8TypeVar.GetSchema()

	assert.Equal(t, expectedSchema, actualSchema, "should return valid open api format that fits uint8s")

	// Test Convert
	var val reflect.Value
	var err error

	val, err = uint8TypeVar.Convert("123")
	assert.Nil(t, err, "should not return error for valid uint8 (123) value")
	assert.Equal(t, uint8(123), val.Interface().(uint8), "should have returned the uint8 value 123")

	val, err = uint8TypeVar.Convert("")
	assert.Nil(t, err, "should not return error for valid uint8 (blank) value")
	assert.Equal(t, uint8(0), val.Interface().(uint8), "should have returned the deafult uint8 value")

	val, err = uint8TypeVar.Convert("not a number")
	assert.Error(t, err, fmt.Errorf(convertError, "not a number"), "should return error for invalid uint8 value (NaN)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid uint8 (NaN)")

	val, err = uint8TypeVar.Convert("-1")
	assert.Error(t, err, fmt.Errorf(convertError, "-1"), "should return error for invalid uint8 value (-1)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid uint8 (-1)")

	tooBig := fmt.Sprint(math.MaxUint8 + 1)
	val, err = uint8TypeVar.Convert(tooBig)
	assert.Error(t, err, fmt.Errorf(convertError, tooBig), "should return error for invalid uint8 value (too large)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid uint8 (too large)")
}

func TestUint16Type(t *testing.T) {
	var uint16TypeVar = new(uint16Type)

	// Test GetSchema
	expectedSchema := spec.Int64Property()
	minimum := float64(0)
	expectedSchema.Minimum = &minimum
	maximum := float64(math.MaxUint16)
	expectedSchema.Maximum = &maximum
	actualSchema := uint16TypeVar.GetSchema()

	assert.Equal(t, expectedSchema, actualSchema, "should return valid open api format that fits uint16s")

	// Test Convert
	var val reflect.Value
	var err error

	val, err = uint16TypeVar.Convert("123")
	assert.Nil(t, err, "should not return error for valid uint16 (123) value")
	assert.Equal(t, uint16(123), val.Interface().(uint16), "should have returned the uint16 value 123")

	val, err = uint16TypeVar.Convert("")
	assert.Nil(t, err, "should not return error for valid uint16 (blank) value")
	assert.Equal(t, uint16(0), val.Interface().(uint16), "should have returned the deafult uint16 value")

	val, err = uint16TypeVar.Convert("not a number")
	assert.Error(t, err, fmt.Errorf(convertError, "not a number"), "should return error for invalid uint16 value (NaN)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid uint16 (NaN)")

	val, err = uint16TypeVar.Convert("-1")
	assert.Error(t, err, fmt.Errorf(convertError, "-1"), "should return error for invalid uint16 value (-1)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid uint16 (-1)")

	tooBig := fmt.Sprint(math.MaxUint16 + 1)
	val, err = uint16TypeVar.Convert(tooBig)
	assert.Error(t, err, fmt.Errorf(convertError, tooBig), "should return error for invalid uint16 value (too large)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid uint16 (too large)")
}

func TestUint32Type(t *testing.T) {
	var uint32TypeVar = new(uint32Type)

	// Test GetSchema
	expectedSchema := spec.Int64Property()
	minimum := float64(0)
	expectedSchema.Minimum = &minimum
	maximum := float64(math.MaxUint32)
	expectedSchema.Maximum = &maximum
	actualSchema := uint32TypeVar.GetSchema()

	assert.Equal(t, expectedSchema, actualSchema, "should return valid open api format that fits uint32s")

	// Test Convert
	var val reflect.Value
	var err error

	val, err = uint32TypeVar.Convert("123")
	assert.Nil(t, err, "should not return error for valid uint32 (123) value")
	assert.Equal(t, uint32(123), val.Interface().(uint32), "should have returned the uint32 value 123")

	val, err = uint32TypeVar.Convert("")
	assert.Nil(t, err, "should not return error for valid uint32 (blank) value")
	assert.Equal(t, uint32(0), val.Interface().(uint32), "should have returned the deafult uint32 value")

	val, err = uint32TypeVar.Convert("not a number")
	assert.Error(t, err, fmt.Errorf(convertError, "not a number"), "should return error for invalid uint32 value (NaN)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid uint32 (NaN)")

	val, err = uint32TypeVar.Convert("-1")
	assert.Error(t, err, fmt.Errorf(convertError, "-1"), "should return error for invalid uint32 value (-1)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid uint32 (-1)")

	tooBig := fmt.Sprint(math.MaxUint32 + 1)
	val, err = uint32TypeVar.Convert(tooBig)
	assert.Error(t, err, fmt.Errorf(convertError, tooBig), "should return error for invalid uint32 value (too large)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid uint32 (too large)")
}

func TestUint64Type(t *testing.T) {
	var uint64TypeVar = new(uint64Type)

	// Test GetSchema
	expectedSchema := spec.Float64Property()
	multOf := float64(1)
	expectedSchema.MultipleOf = &multOf
	minimum := float64(0)
	expectedSchema.Minimum = &minimum
	maximum := float64(math.MaxUint64)
	expectedSchema.Maximum = &maximum
	actualSchema := uint64TypeVar.GetSchema()

	assert.Equal(t, expectedSchema, actualSchema, "should return valid open api format that fits uint64s")

	// Test Convert
	var val reflect.Value
	var err error

	val, err = uint64TypeVar.Convert("123")
	assert.Nil(t, err, "should not return error for valid uint64 (123) value")
	assert.Equal(t, uint64(123), val.Interface().(uint64), "should have returned the uint64 value 123")

	val, err = uint64TypeVar.Convert("")
	assert.Nil(t, err, "should not return error for valid uint64 (blank) value")
	assert.Equal(t, uint64(0), val.Interface().(uint64), "should have returned the deafult uint64 value")

	val, err = uint64TypeVar.Convert("not a number")
	assert.Error(t, err, fmt.Errorf(convertError, "not a number"), "should return error for invalid uint64 value (NaN)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid uint64 (NaN)")

	val, err = uint64TypeVar.Convert("-1")
	assert.Error(t, err, fmt.Errorf(convertError, "-1"), "should return error for invalid uint64 value (-1)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid uint64 (-1)")
}

func TestFloat32Type(t *testing.T) {
	var float32TypeVar = new(float32Type)

	// Test GetSchema
	assert.Equal(t, spec.Float32Property(), float32TypeVar.GetSchema(), "should return open api float32 spec")

	// Test Convert
	var val reflect.Value
	var err error

	val, err = float32TypeVar.Convert("123")
	assert.Nil(t, err, "should not return error for valid float32 (123) value")
	assert.Equal(t, float32(123), val.Interface().(float32), "should have returned the float32 value 123")

	val, err = float32TypeVar.Convert("123.456")
	assert.Nil(t, err, "should not return error for valid float32 (123.456) value")
	assert.Equal(t, float32(123.456), val.Interface().(float32), "should have returned the float32 value 123.456")

	val, err = float32TypeVar.Convert("")
	assert.Nil(t, err, "should not return error for valid float32 (blank) value")
	assert.Equal(t, float32(0), val.Interface().(float32), "should have returned the deafult float32 value")

	val, err = float32TypeVar.Convert("not a number")
	assert.Error(t, err, fmt.Errorf(convertError, "not a number"), "should return error for invalid float32 value (NaN)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid float32 (NaN)")

	tooBig := fmt.Sprint(math.MaxFloat64)
	val, err = float32TypeVar.Convert(tooBig)
	assert.Error(t, err, fmt.Errorf(convertError, tooBig), "should return error for invalid uint32 value (too large)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid uint32 (too large)")
}

func TestFloat64Type(t *testing.T) {
	var float64TypeVar = new(float64Type)

	// Test GetSchema
	assert.Equal(t, spec.Float64Property(), float64TypeVar.GetSchema(), "should return open api float64 spec")

	// Test Convert
	var val reflect.Value
	var err error

	val, err = float64TypeVar.Convert("123")
	assert.Nil(t, err, "should not return error for valid float64 (123) value")
	assert.Equal(t, float64(123), val.Interface().(float64), "should have returned the float64 value 123")

	val, err = float64TypeVar.Convert("123.456")
	assert.Nil(t, err, "should not return error for valid float64 (123.456) value")
	assert.Equal(t, float64(123.456), val.Interface().(float64), "should have returned the float64 value 123.456")

	val, err = float64TypeVar.Convert("")
	assert.Nil(t, err, "should not return error for valid float64 (blank) value")
	assert.Equal(t, float64(0), val.Interface().(float64), "should have returned the deafult float64 value")

	val, err = float64TypeVar.Convert("not a number")
	assert.Error(t, err, fmt.Errorf(convertError, "not a number"), "should return error for invalid float64 value (NaN)")
	assert.Equal(t, reflect.Value{}, val, "should have returned the blank value for invalid float64 (NaN)")
}

func TestInterfaceType(t *testing.T) {
	var interfaceTypeVar = new(interfaceType)

	// Test GetSchema
	assert.Equal(t, new(spec.Schema), interfaceTypeVar.GetSchema(), "should return open api float64 spec")

	// Test Convert
	var val reflect.Value
	var err error

	val, err = interfaceTypeVar.Convert("hello world")
	assert.Nil(t, err, "should never return error for interface")
	assert.Equal(t, "hello world", val.Interface().(string), "should return string that went in")
}
