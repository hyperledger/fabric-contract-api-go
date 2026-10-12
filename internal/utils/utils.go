// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// ValidateErrorsToString converts errors from JSON schema output into readable string
func ValidateErrorsToString(err error) string {
	var valErr *jsonschema.ValidationError
	if !errors.As(err, &valErr) {
		return err.Error()
	}

	output := valErr.BasicOutput()
	hasSpecificErrors := false
	for _, u := range output.Errors {
		if u.Error != nil && u.Error.String() != "validation failed" {
			hasSpecificErrors = true
			break
		}
	}

	msgs := make([]string, 0, len(output.Errors))
	for _, u := range output.Errors {
		errStr := "validation failed"
		if u.Error != nil {
			errStr = u.Error.String()
		}
		if hasSpecificErrors && errStr == "validation failed" {
			continue
		}
		loc := strings.TrimPrefix(u.InstanceLocation, "/")
		msg := ""
		if loc != "" {
			msg = loc + ": " + errStr
		} else {
			msg = "(root): " + errStr
		}
		msgs = append(msgs, msg)
	}

	if len(msgs) == 0 {
		return valErr.Error()
	}

	sort.Strings(msgs)

	var sb strings.Builder
	for i, v := range msgs {
		sb.WriteString(strconv.Itoa(i + 1))
		sb.WriteString(". ")
		sb.WriteString(v)
		sb.WriteByte('\n')
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// StringInSlice returns whether string exists in string slice
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// SliceAsCommaSentence returns string slice as comma separated sentence
func SliceAsCommaSentence(slice []string) string {
	return strings.Replace(strings.Join(slice, " and "), " and ", ", ", len(slice)-2)
}
