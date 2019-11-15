// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

// ================================
// HELPERS
// ================================

type MyResultError struct {
	gojsonschema.ResultError
	message string
}

func (re MyResultError) String() string {
	return re.message
}

// ================================
// TESTS
// ================================

func TestValidateErrorsToString(t *testing.T) {
	// should join errors with a new line
	error1 := MyResultError{
		message: "some error message",
	}
	error2 := MyResultError{
		message: "yet another error message",
	}

	assert.Equal(t, "1. some error message", ValidateErrorsToString([]gojsonschema.ResultError{error1}), "should return nicely formatted single error")
	assert.Equal(t, "1. some error message\n2. yet another error message", ValidateErrorsToString([]gojsonschema.ResultError{error1, error2}), "should return nicely formatted multiple error")
}

func TestStringInSlice(t *testing.T) {
	slice := []string{"word", "another word"}

	// Should return true when string present in slice
	assert.True(t, StringInSlice("word", slice), "should have returned true when sling in slice")

	// Should return false when string no present in slice
	assert.False(t, StringInSlice("bad word", slice), "should have returned true when sling in slice")
}
