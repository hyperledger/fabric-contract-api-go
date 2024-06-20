// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package metadata

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadLocalFile(t *testing.T) {

	file, err := readLocalFile("i don't exist")
	_, expectedErr := os.ReadFile("i don't exist")
	assert.Nil(t, file, "should not return file on error")
	assert.Contains(t, err.Error(), strings.Split(expectedErr.Error(), ":")[1], "should return same error as ioutils read file")

	file, err = readLocalFile("schema/schema.json")
	expectedFile, _ := os.ReadFile("./schema/schema.json")
	assert.Equal(t, expectedFile, file, "should return same file")
	assert.Nil(t, err, "should return same err")
}
