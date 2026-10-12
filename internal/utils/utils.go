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
		if u.Error.String() != "validation failed" {
			hasSpecificErrors = true
			break
		}
	}

	msgs := make([]string, 0, len(output.Errors))
	for _, u := range output.Errors {
		if hasSpecificErrors && u.Error.String() == "validation failed" {
			continue
		}
		loc := strings.TrimPrefix(u.InstanceLocation, "/")
		msg := ""
		if loc != "" {
			msg = loc + ": " + u.Error.String()
		} else {
			msg = "(root): " + u.Error.String()
		}
		msgs = append(msgs, msg)
	}

	sort.Strings(msgs)

	toReturn := ""
	for i, v := range msgs {
		toReturn += strconv.Itoa(i+1) + ". " + v + "\n"
	}

	return strings.Trim(toReturn, "\n")
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
