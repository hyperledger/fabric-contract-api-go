// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-contract-api-go/internal/types"
	"github.com/hyperledger/fabric-contract-api-go/internal/utils"
	"github.com/stretchr/testify/assert"
)

// ================================
// HELPERS
// ================================
const basicErr = "Type %s is not valid. Expected a struct or one of the basic types %s or an array/slice of these"

type goodStruct struct {
	Prop1 string
	Prop2 int `json:"prop2"`
}

type BadStruct struct {
	Prop1 string    `json:"Prop1"`
	Prop2 complex64 `json:"prop2"`
}

type UsefulInterface interface{}

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

type myInterface interface {
	SomeFunction(string, int) (string, error)
}

type structFailsParamLength struct{}

func (s *structFailsParamLength) SomeFunction(param1 string) (string, error) {
	return "", nil
}

type structFailsParamType struct{}

func (s *structFailsParamType) SomeFunction(param1 string, param2 float32) (string, error) {
	return "", nil
}

type structFailsReturnLength struct{}

func (s *structFailsReturnLength) SomeFunction(param1 string, param2 int) string {
	return ""
}

type structFailsReturnType struct{}

func (s *structFailsReturnType) SomeFunction(param1 string, param2 int) (string, int) {
	return "", 0
}

type structMeetsInterface struct{}

func (s *structMeetsInterface) SomeFunction(param1 string, param2 int) (string, error) {
	return "", nil
}

// ================================
// TESTS
// ================================

func TestListBasicTypes(t *testing.T) {
	types := []string{"bool", "float32", "float64", "int", "int16", "int32", "int64", "int8", "interface", "string", "uint", "uint16", "uint32", "uint64", "uint8"}

	assert.Equal(t, utils.SliceAsCommaSentence(types), listBasicTypes(), "should return basic types as a human readable list")
}

func TestArrayOfValidType(t *testing.T) {
	// Further tested by typeIsValid array tests

	var err error

	zeroArr := [0]int{}
	err = arrayOfValidType(reflect.ValueOf(zeroArr), []reflect.Type{})
	assert.Equal(t, errors.New("Arrays must have length greater than 0"), err, "should throw error when 0 length array passed")

	badArr := [1]complex128{}
	err = arrayOfValidType(reflect.ValueOf(badArr), []reflect.Type{})
	assert.EqualError(t, err, typeIsValid(reflect.TypeOf(complex128(1)), []reflect.Type{}, false).Error(), "should throw error when invalid type passed")
}

func TestStructOfValidType(t *testing.T) {
	assert.Nil(t, structOfValidType(reflect.TypeOf(new(goodStruct)), []reflect.Type{}), "should not return an error for a pointer struct")

	assert.Nil(t, structOfValidType(reflect.TypeOf(goodStruct{}), []reflect.Type{}), "should not return an error for a valid struct")

	assert.EqualError(t, structOfValidType(reflect.TypeOf(BadStruct{}), []reflect.Type{}), fmt.Sprintf(basicErr, badType.String(), listBasicTypes()), "should return an error for invalid struct")
}

func TestTypeIsValid(t *testing.T) {
	// HELPERS
	badArr := reflect.New(badArrayType).Elem()

	type goodStruct2 struct {
		Prop1 goodStruct
	}

	type goodStruct3 struct {
		Prop1 *goodStruct
	}

	type goodStruct4 struct {
		Prop1 interface{}
	}

	type goodStruct5 struct {
		Prop1 *goodStruct5
	}

	type BadStruct2 struct {
		Prop1 BadStruct
	}

	type BadStruct3 struct {
		Prop1 UsefulInterface
	}

	// TESTS
	assert.Nil(t, typeIsValid(boolRefType, []reflect.Type{}, false), "should not return an error for a bool type")
	assert.Nil(t, typeIsValid(stringRefType, []reflect.Type{}, false), "should not return an error for a string type")
	assert.Nil(t, typeIsValid(intRefType, []reflect.Type{}, false), "should not return an error for int type")
	assert.Nil(t, typeIsValid(int8RefType, []reflect.Type{}, false), "should not return an error for int8 type")
	assert.Nil(t, typeIsValid(int16RefType, []reflect.Type{}, false), "should not return an error for int16 type")
	assert.Nil(t, typeIsValid(int32RefType, []reflect.Type{}, false), "should not return an error for int32 type")
	assert.Nil(t, typeIsValid(int64RefType, []reflect.Type{}, false), "should not return an error for int64 type")
	assert.Nil(t, typeIsValid(uintRefType, []reflect.Type{}, false), "should not return an error for uint type")
	assert.Nil(t, typeIsValid(uint8RefType, []reflect.Type{}, false), "should not return an error for uint8 type")
	assert.Nil(t, typeIsValid(uint16RefType, []reflect.Type{}, false), "should not return an error for uint16 type")
	assert.Nil(t, typeIsValid(uint32RefType, []reflect.Type{}, false), "should not return an error for uint32 type")
	assert.Nil(t, typeIsValid(uint64RefType, []reflect.Type{}, false), "should not return an error for uint64 type")
	assert.Nil(t, typeIsValid(float32RefType, []reflect.Type{}, false), "should not return an error for float32 type")
	assert.Nil(t, typeIsValid(float64RefType, []reflect.Type{}, false), "should not return an error for float64 type")
	assert.Nil(t, typeIsValid(float64RefType, []reflect.Type{}, false), "should not return an error for float64 type")
	assert.Nil(t, typeIsValid(reflect.TypeOf(goodStruct4{}).Field(0).Type, []reflect.Type{}, false), "should not return error for interface{} type")

	assert.Nil(t, typeIsValid(types.ErrorType, []reflect.Type{}, true), "should not return an error for error type on allow error")
	assert.Nil(t, typeIsValid(types.TimeType, []reflect.Type{}, false), "should not return an error for time type on allow error")

	assert.Nil(t, typeIsValid(reflect.TypeOf([1]string{}), []reflect.Type{}, false), "should not return an error for a string array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1]bool{}), []reflect.Type{}, false), "should not return an error for a bool array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1]int{}), []reflect.Type{}, false), "should not return an error for an int array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1]int8{}), []reflect.Type{}, false), "should not return an error for an int8 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1]int16{}), []reflect.Type{}, false), "should not return an error for an int16 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1]int32{}), []reflect.Type{}, false), "should not return an error for an int32 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1]int64{}), []reflect.Type{}, false), "should not return an error for an int64 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1]uint{}), []reflect.Type{}, false), "should not return an error for a uint array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1]uint8{}), []reflect.Type{}, false), "should not return an error for a uint8 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1]uint16{}), []reflect.Type{}, false), "should not return an error for a uint16 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1]uint32{}), []reflect.Type{}, false), "should not return an error for a uint32 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1]uint64{}), []reflect.Type{}, false), "should not return an error for a uint64 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1]float32{}), []reflect.Type{}, false), "should not return an error for a float32 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1]float64{}), []reflect.Type{}, false), "should not return an error for a float64 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1]byte{}), []reflect.Type{}, false), "should not return an error for a float64 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1]rune{}), []reflect.Type{}, false), "should not return an error for a float64 array type")

	assert.Nil(t, typeIsValid(reflect.TypeOf([1][1]string{}), []reflect.Type{}, false), "should not return an error for a multidimensional string array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1][1]bool{}), []reflect.Type{}, false), "should not return an error for a multidimensional bool array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1][1]int{}), []reflect.Type{}, false), "should not return an error for an multidimensional int array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1][1]int8{}), []reflect.Type{}, false), "should not return an error for an multidimensional int8 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1][1]int16{}), []reflect.Type{}, false), "should not return an error for an multidimensional int16 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1][1]int32{}), []reflect.Type{}, false), "should not return an error for an multidimensional int32 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1][1]int64{}), []reflect.Type{}, false), "should not return an error for an multidimensional int64 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1][1]uint{}), []reflect.Type{}, false), "should not return an error for a multidimensional uint array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1][1]uint8{}), []reflect.Type{}, false), "should not return an error for a multidimensional uint8 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1][1]uint16{}), []reflect.Type{}, false), "should not return an error for a multidimensional uint16 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1][1]uint32{}), []reflect.Type{}, false), "should not return an error for a multidimensional uint32 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1][1]uint64{}), []reflect.Type{}, false), "should not return an error for a multidimensional uint64 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1][1]float32{}), []reflect.Type{}, false), "should not return an error for a multidimensional float32 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1][1]float64{}), []reflect.Type{}, false), "should not return an error for a multidimensional float64 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1][1]byte{}), []reflect.Type{}, false), "should not return an error for a multidimensional float64 array type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([1][1]rune{}), []reflect.Type{}, false), "should not return an error for a multidimensional float64 array type")

	assert.Nil(t, typeIsValid(reflect.TypeOf([1][2][3][4][5][6][7][8]string{}), []reflect.Type{}, false), "should not return an error for a very multidimensional string array type")

	assert.Nil(t, typeIsValid(reflect.TypeOf([]string{}), []reflect.Type{}, false), "should not return an error for a string slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([]bool{}), []reflect.Type{}, false), "should not return an error for a bool slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([]int{}), []reflect.Type{}, false), "should not return an error for a int slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([]int8{}), []reflect.Type{}, false), "should not return an error for a int8 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([]int16{}), []reflect.Type{}, false), "should not return an error for a int16 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([]int32{}), []reflect.Type{}, false), "should not return an error for a int32 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([]int64{}), []reflect.Type{}, false), "should not return an error for a int64 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([]uint{}), []reflect.Type{}, false), "should not return an error for a uint slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([]uint8{}), []reflect.Type{}, false), "should not return an error for a uint8 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([]uint16{}), []reflect.Type{}, false), "should not return an error for a uint16 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([]uint32{}), []reflect.Type{}, false), "should not return an error for a uint32 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([]uint64{}), []reflect.Type{}, false), "should not return an error for a uint64 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([]float32{}), []reflect.Type{}, false), "should not return an error for a float32 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([]float64{}), []reflect.Type{}, false), "should not return an error for a float64 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([]byte{}), []reflect.Type{}, false), "should not return an error for a byte slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([]rune{}), []reflect.Type{}, false), "should not return an error for a rune slice type")

	assert.Nil(t, typeIsValid(reflect.TypeOf([][]string{}), []reflect.Type{}, false), "should not return an error for a multidimensional string slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([][]bool{}), []reflect.Type{}, false), "should not return an error for a multidimensional bool slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([][]int{}), []reflect.Type{}, false), "should not return an error for a multidimensional int slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([][]int8{}), []reflect.Type{}, false), "should not return an error for a multidimensional int8 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([][]int16{}), []reflect.Type{}, false), "should not return an error for a multidimensional int16 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([][]int32{}), []reflect.Type{}, false), "should not return an error for a multidimensional int32 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([][]int64{}), []reflect.Type{}, false), "should not return an error for a multidimensional int64 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([][]uint{}), []reflect.Type{}, false), "should not return an error for a multidimensional uint slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([][]uint8{}), []reflect.Type{}, false), "should not return an error for a multidimensional uint8 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([][]uint16{}), []reflect.Type{}, false), "should not return an error for a multidimensional uint16 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([][]uint32{}), []reflect.Type{}, false), "should not return an error for a multidimensional uint32 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([][]uint64{}), []reflect.Type{}, false), "should not return an error for a multidimensional uint64 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([][]float32{}), []reflect.Type{}, false), "should not return an error for a multidimensional float32 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([][]float64{}), []reflect.Type{}, false), "should not return an error for a multidimensional float64 slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([][]byte{}), []reflect.Type{}, false), "should not return an error for a multidimensional byte slice type")
	assert.Nil(t, typeIsValid(reflect.TypeOf([][]rune{}), []reflect.Type{}, false), "should not return an error for a multidimensional rune slice type")

	assert.Nil(t, typeIsValid(reflect.TypeOf([][][][][][][][]string{}), []reflect.Type{}, false), "should not return an error for a very multidimensional string slice type")

	assert.Nil(t, typeIsValid(reflect.TypeOf([2][]string{}), []reflect.Type{}, false), "should not return an error for a string slice of array type")

	assert.Nil(t, typeIsValid(reflect.TypeOf(goodStruct{}), []reflect.Type{}, false), "should not return an error for a valid struct")

	assert.Nil(t, typeIsValid(reflect.TypeOf([1]goodStruct{}), []reflect.Type{}, false), "should not return an error for an array of valid struct")

	assert.Nil(t, typeIsValid(reflect.TypeOf([]goodStruct{}), []reflect.Type{}, false), "should not return an error for a slice of valid struct")

	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string]string{}), []reflect.Type{}, false), "should not return an error for a map string item type")
	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string]bool{}), []reflect.Type{}, false), "should not return an error for a map bool item type")
	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string]int{}), []reflect.Type{}, false), "should not return an error for a map int item type")
	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string]int8{}), []reflect.Type{}, false), "should not return an error for a map int8 item type")
	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string]int16{}), []reflect.Type{}, false), "should not return an error for a map int16 item type")
	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string]int32{}), []reflect.Type{}, false), "should not return an error for a map int32 item type")
	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string]int64{}), []reflect.Type{}, false), "should not return an error for a map int64 item type")
	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string]uint{}), []reflect.Type{}, false), "should not return an error for a map uint item type")
	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string]uint8{}), []reflect.Type{}, false), "should not return an error for a map uint8 item type")
	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string]uint16{}), []reflect.Type{}, false), "should not return an error for a map uint16 item type")
	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string]uint32{}), []reflect.Type{}, false), "should not return an error for a map uint32 item type")
	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string]uint64{}), []reflect.Type{}, false), "should not return an error for a map uint64 item type")
	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string]float32{}), []reflect.Type{}, false), "should not return an error for a map float32 item type")
	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string]float64{}), []reflect.Type{}, false), "should not return an error for a map float64 item type")
	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string]byte{}), []reflect.Type{}, false), "should not return an error for a map byte item type")
	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string]rune{}), []reflect.Type{}, false), "should not return an error for a map rune item type")

	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string]map[string]string{}), []reflect.Type{}, false), "should not return an error for a map of map")

	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string]goodStruct{}), []reflect.Type{}, false), "should not return an error for a map with struct item type")

	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string][1]string{}), []reflect.Type{}, false), "should not return an error for a map with string array item type")

	assert.Nil(t, typeIsValid(reflect.TypeOf(map[string][]string{}), []reflect.Type{}, false), "should not return an error for a map with string slice item type")

	assert.Nil(t, typeIsValid(reflect.TypeOf(goodStruct2{}), []reflect.Type{}, false), "should not return an error for a valid struct with struct property")

	assert.Nil(t, typeIsValid(reflect.TypeOf(goodStruct3{}), []reflect.Type{}, false), "should not return an error for a valid struct with struct ptr property")

	assert.Nil(t, typeIsValid(reflect.TypeOf(goodStruct5{}), []reflect.Type{}, false), "should not return an error for a valid struct with cyclic dependency")

	assert.Nil(t, typeIsValid(badType, []reflect.Type{badType}, false), "should not error when type not in basic types but is in additional types")
	assert.Nil(t, typeIsValid(reflect.TypeOf(BadStruct{}), []reflect.Type{reflect.TypeOf(BadStruct{})}, false), "should not error when bad struct is in additional types")
	assert.Nil(t, typeIsValid(reflect.TypeOf(BadStruct2{}), []reflect.Type{reflect.TypeOf(BadStruct{})}, false), "should not error when bad struct is in additional types and passed type has that as property")

	assert.EqualError(t, typeIsValid(badType, []reflect.Type{}, false), fmt.Sprintf(basicErr, badType.String(), listBasicTypes()), "should have returned error for invalid basic type")

	assert.EqualError(t, typeIsValid(badArrayType, []reflect.Type{}, false), arrayOfValidType(badArr, []reflect.Type{}).Error(), "should have returned error for invalid array type")

	assert.EqualError(t, typeIsValid(badSliceType, []reflect.Type{}, false), fmt.Sprintf(basicErr, badType.String(), listBasicTypes()), "should have returned error for invalid slice type")

	assert.EqualError(t, typeIsValid(badMapItemType, []reflect.Type{}, false), fmt.Sprintf(basicErr, badType.String(), listBasicTypes()), "should have returned error for invalid map item type")

	assert.EqualError(t, typeIsValid(badMapKeyType, []reflect.Type{}, false), "Map key type complex64 is not valid. Expected string", "should have returned error for invalid map key type")

	zeroMultiArr := [1][0]int{}
	err := typeIsValid(reflect.TypeOf(zeroMultiArr), []reflect.Type{}, false)
	assert.Equal(t, errors.New("Arrays must have length greater than 0"), err, "should throw error when 0 length array passed in multi level array")

	err = typeIsValid(types.ErrorType, []reflect.Type{}, false)
	assert.EqualError(t, err, fmt.Sprintf(basicErr, types.ErrorType.String(), listBasicTypes()), "should throw error when error passed and allowError false")

	badMultiArr := [1][1]complex128{}
	err = typeIsValid(reflect.TypeOf(badMultiArr), []reflect.Type{}, false)
	assert.Equal(t, fmt.Errorf(basicErr, "complex128", listBasicTypes()), err, "should throw error when bad multidimensional array passed")

	badMultiSlice := [][]complex128{}
	err = typeIsValid(reflect.TypeOf(badMultiSlice), []reflect.Type{}, false)
	assert.Equal(t, fmt.Errorf(basicErr, "complex128", listBasicTypes()), err, "should throw error when 0 length array passed")

	assert.EqualError(t, typeIsValid(reflect.TypeOf([]BadStruct{}), []reflect.Type{}, false), fmt.Sprintf(basicErr, badType.String(), listBasicTypes()), "should return an error for array of invalid struct")

	assert.EqualError(t, typeIsValid(reflect.TypeOf([]BadStruct{}), []reflect.Type{}, false), fmt.Sprintf(basicErr, badType.String(), listBasicTypes()), "should return an error for slice of invalid struct")

	assert.EqualError(t, typeIsValid(reflect.TypeOf(BadStruct2{}), []reflect.Type{}, false), fmt.Sprintf(basicErr, badType.String(), listBasicTypes()), "should return an error for struct with invalid property of a struct")

	assert.EqualError(t, typeIsValid(reflect.TypeOf(BadStruct2{}), []reflect.Type{}, false), fmt.Sprintf(basicErr, badType.String(), listBasicTypes()), "should return an error for struct with invalid property of a pointer to struct")

	assert.EqualError(t, typeIsValid(reflect.TypeOf(BadStruct3{}), []reflect.Type{}, false), fmt.Sprintf(basicErr, "internal.UsefulInterface", listBasicTypes()), "should return an error for struct with invalid property of an interface not (interface{})")

	assert.EqualError(t, typeIsValid(badArrayType, []reflect.Type{badArrayType}, false), arrayOfValidType(badArr, []reflect.Type{badArrayType}).Error(), "should have returned error for invalid array type")

	assert.EqualError(t, typeIsValid(badSliceType, []reflect.Type{badSliceType}, false), fmt.Sprintf(basicErr, badType.String(), listBasicTypes()), "should have returned error for invalid slice type")

	assert.EqualError(t, typeIsValid(badType, []reflect.Type{}, true), fmt.Sprintf(strings.Replace(basicErr, "types", "types%s", 1), badType.String(), " error,", listBasicTypes()), "should have returned include error in list of valid types")
}

func TestTypeMatchesInterface(t *testing.T) {
	var err error

	interfaceType := reflect.TypeOf((*myInterface)(nil)).Elem()

	err = typeMatchesInterface(reflect.TypeOf(new(BadStruct)), reflect.TypeOf(""))
	assert.EqualError(t, err, "Type passed for interface is not an interface", "should error when type passed is not an interface")

	err = typeMatchesInterface(reflect.TypeOf(new(BadStruct)), interfaceType)
	assert.EqualError(t, err, "Missing function SomeFunction", "should error when type passed is missing required method in interface")

	err = typeMatchesInterface(reflect.TypeOf(new(structFailsParamLength)), interfaceType)
	assert.EqualError(t, err, "Parameter mismatch in method SomeFunction. Expected 2, got 1", "should error when type passed has method but different number of parameters")

	err = typeMatchesInterface(reflect.TypeOf(new(structFailsParamType)), interfaceType)
	assert.EqualError(t, err, "Parameter mismatch in method SomeFunction at parameter 1. Expected int, got float32", "should error when type passed has method but different parameter types")

	err = typeMatchesInterface(reflect.TypeOf(new(structFailsReturnLength)), interfaceType)
	assert.EqualError(t, err, "Return mismatch in method SomeFunction. Expected 2, got 1", "should error when type passed has method but different number of returns")

	err = typeMatchesInterface(reflect.TypeOf(new(structFailsReturnType)), interfaceType)
	assert.EqualError(t, err, "Return mismatch in method SomeFunction at return 1. Expected error, got int", "should error when type passed has method but different return types")

	err = typeMatchesInterface(reflect.TypeOf(new(structMeetsInterface)), interfaceType)
	assert.Nil(t, err, "should not error when struct meets interface")
}
