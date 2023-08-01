// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package field

import (
	"errors"
	"fmt"
	"reflect"
	"unicode"
)

func SetValue(s any, fieldName string, fieldValue any) error {
	// Get the reflect.Value of the struct.
	v := reflect.ValueOf(s)

	if fieldName == "" {
		return errors.New("setValue: fieldName can not be empty")
	}

	if unicode.IsLower(rune(fieldName[0])) {
		return errors.New("setValue: fieldName can not be unexported")
	}

	// Check if the input is a pointer.
	if v.Kind() != reflect.Ptr {
		return errors.New("setValue: input is not a pointer")
	}

	// Dereference the pointer to get the underlying struct.
	v = v.Elem()

	// If the input is not a struct, return an error.
	if v.Kind() != reflect.Struct {
		return errors.New("setValue: input is not a struct")
	}

	// Iterate over all fields of the struct.
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if unicode.IsLower(rune(v.Type().Field(i).Name[0])) {
			continue
		}

		// If the field is a struct, recursively call the setFieldValue function.
		if field.Kind() == reflect.Struct {
			if err := SetValue(field.Addr().Interface(), fieldName, fieldValue); err != nil {
				return err
			}
		}
		// If the field is a pointer, recursively call the setFieldValue function.
		if field.Kind() == reflect.Pointer && !field.IsNil() && field.Elem().Kind() == reflect.Struct {
			if err := SetValue(field.Interface(), fieldName, fieldValue); err != nil {
				return err
			}
		}
		// If the field is a array, recursively call the setFieldValue function.
		if field.Kind() == reflect.Slice {
			for i := 0; i < field.Len(); i++ {
				elem := field.Index(i)
				// If the field is a struct, recursively call the setValue function.
				if elem.Kind() == reflect.Struct {
					if err := SetValue(elem.Addr().Interface(), fieldName, fieldValue); err != nil {
						return err
					}
				}
				// If the field is a pointer, recursively call the setValue function.
				if elem.Kind() == reflect.Pointer && !elem.IsNil() && elem.Elem().Kind() == reflect.Struct {
					if err := SetValue(elem.Interface(), fieldName, fieldValue); err != nil {
						return err
					}
				}
			}
		}

		// If the field is a map, recursively call the setValue function
		if field.Kind() == reflect.Map {
			for _, k := range field.MapKeys() {
				v := field.MapIndex(k)
				if v.Kind() == reflect.Pointer && !v.IsNil() && v.Elem().Kind() == reflect.Struct {
					if err := SetValue(v.Interface(), fieldName, fieldValue); err != nil {
						return err
					}
				}
			}
		}

		// If the field name matches, set the field value.
		if v.Type().Field(i).Name == fieldName {
			if !field.CanSet() {
				return fmt.Errorf("setValue: field %s is not settable", fieldName)
			}
			fieldValueReflect := reflect.ValueOf(fieldValue)
			if fieldValueReflect.Type().AssignableTo(field.Type()) {
				field.Set(fieldValueReflect)
			} else {
				return fmt.Errorf("setValue: field %s type mismatch", fieldName)
			}
		}
	}

	return nil
}
