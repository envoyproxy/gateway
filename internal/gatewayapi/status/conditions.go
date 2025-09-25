// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from Contour,
// https://github.com/projectcontour/contour
// from the source file
// https://github.com/projectcontour/contour/blob/main/internal/status/gatewayclassconditions.go
// and is provided here subject to the following:
// Copyright Project Contour Authors
// SPDX-License-Identifier: Apache-2.0

package status

import (
	"time"
	"unicode"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MergeConditions adds or updates matching conditions, and updates the transition
// time if details of a condition have changed. Returns the updated condition array.
func MergeConditions(conditions []metav1.Condition, updates ...metav1.Condition) []metav1.Condition {
	var additions []metav1.Condition
	for i, update := range updates {
		add := true
		for j, cond := range conditions {
			if cond.Type == update.Type {
				add = false
				if conditionChanged(&cond, &update) {
					conditions[j].Status = update.Status
					conditions[j].Reason = update.Reason
					conditions[j].Message = update.Message
					conditions[j].ObservedGeneration = update.ObservedGeneration
					conditions[j].LastTransitionTime = update.LastTransitionTime
					break
				}
			}
		}
		if add {
			additions = append(additions, updates[i])
		}
	}
	conditions = append(conditions, additions...)
	return conditions
}

func newCondition(t string, status metav1.ConditionStatus, reason, msg string, lt time.Time, og int64) metav1.Condition {
	return metav1.Condition{
		Type:               t,
		Status:             status,
		Reason:             reason,
		Message:            msg,
		LastTransitionTime: metav1.NewTime(lt),
		ObservedGeneration: og,
	}
}

func conditionChanged(a, b *metav1.Condition) bool {
	opts := cmpopts.IgnoreFields(metav1.Condition{}, "Type", "LastTransitionTime")
	return !cmp.Equal(*a, *b, opts)
}

// Error2ConditionMsg format the error string to a Status condition message.
// * Convert the first letter to capital
// * Append "." to the string if it doesn't exist
func Error2ConditionMsg(err error) string {
	if err == nil {
		return ""
	}

	message := err.Error()
	if message == "" {
		return message
	}

	// Convert the string to a rune slice for easier manipulation
	runes := []rune(message)

	// Check if the first rune is a letter and convert it to uppercase
	if unicode.IsLetter(runes[0]) {
		runes[0] = unicode.ToUpper(runes[0])
	}

	// Check if the last rune is a punctuation '.' and append it if not
	last := runes[len(runes)-1]
	if !unicode.IsPunct(last) || last != '.' {
		runes = append(runes, '.')
	}

	// Convert the rune slice back to a string
	return string(runes)
}
