// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package contractapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ================================
// Tests
// ================================

func TestSetMetadata(t *testing.T) {
	sc := SystemContract{}
	sc.setMetadata("my metadata")

	assert.Equal(t, "my metadata", sc.metadata, "should have set metadata field")
}

func TestGetMetadata(t *testing.T) {
	sc := SystemContract{}
	sc.metadata = "my metadata"

	assert.Equal(t, "my metadata", sc.GetMetadata(), "should have returned metadata field")
}

func TestGetEvaluateTransactions(t *testing.T) {
	sc := SystemContract{}

	assert.Equal(t, []string{"GetMetadata"}, sc.GetEvaluateTransactions(), "should have returned functions names that should be evaluate")
}
