// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestValidateErrorsToString(t *testing.T) {
	// Non-validation error fallback
	genericErr := errors.New("generic error")
	assert.Equal(t, "generic error", ValidateErrorsToString(genericErr))

	// Single root error
	c := jsonschema.NewCompiler()
	_ = c.AddResource("schema1.json", map[string]any{
		"type": "string",
	})
	sch1, err := c.Compile("schema1.json")
	require.NoError(t, err)

	valErr1 := sch1.Validate(123)
	assert.Equal(t, "1. (root): got number, want string", ValidateErrorsToString(valErr1))

	// Multiple errors including a property location
	_ = c.AddResource("schema2.json", map[string]any{
		"type":     "object",
		"required": []any{"name"},
		"properties": map[string]any{
			"age": map[string]any{"type": "integer"},
		},
	})
	sch2, err := c.Compile("schema2.json")
	require.NoError(t, err)

	valErr2 := sch2.Validate(map[string]any{
		"age": "invalid",
	})
	assert.Equal(t, "1. (root): missing property 'name'\n2. age: got string, want integer", ValidateErrorsToString(valErr2))
}

func TestStringInSlice(t *testing.T) {
	slice := []string{"word", "another word"}

	// Should return true when string present in slice
	assert.True(t, StringInSlice("word", slice), "should have returned true when string in slice")

	// Should return false when string not present in slice
	assert.False(t, StringInSlice("bad word", slice), "should have returned true when string in slice")
}

func TestSliceAsCommaSentence(t *testing.T) {
	slice := []string{"one", "two", "three"}

	assert.Equal(t, "one, two and three", SliceAsCommaSentence(slice), "should have put commas between slice elements and join last element with and")

	assert.Equal(t, "one", SliceAsCommaSentence([]string{"one"}), "should handle single item")
}
