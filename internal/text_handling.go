// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal

import "strings"

func sliceAsCommaSentence(slice []string) string {
	return strings.Replace(strings.Join(slice, " and "), " and ", ", ", len(slice)-2)
}
