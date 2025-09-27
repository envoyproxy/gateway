// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package jsonpatch

import (
	"reflect"
	"strings"

	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"github.com/pkg/errors"
)

func ConvertPathToPointers(jsonDoc []byte, jsonPath, path string) ([]string, error) {
	jsonPointers := make([]string, 0, 4) // reasonable default for most json path queries

	jObj, err := oj.Parse(jsonDoc)
	if err != nil {
		return nil, errors.Wrap(err, "Error during parsing json")
	}

	jPath, err := jp.ParseString(jsonPath)
	if err != nil {
		return nil, errors.Wrap(err, "Error during parsing jpath")
	}

	if len(jPath) == 1 {
		_, isRoot := jPath[0].(jp.Root)
		if isRoot {
			return nil, errors.New("Using only Root ('$') in json path expression is not allowed!")
		}
	}

	locations := jPath.Locate(jObj, 0)
	for _, l := range locations {
		jsonPointer, err := expToPointer(l)
		if err != nil {
			return nil, errors.Wrap(err, "Error during converting path to pointer")
		}
		jsonPointers = append(jsonPointers, concat(jsonPointer, path))
	}
	return jsonPointers, nil
}

func concat(jsonPointer, path string) string {
	if path == "" {
		return jsonPointer
	}
	const separator string = "/"
	parts := []string{
		strings.TrimSuffix(jsonPointer, separator),
		strings.TrimPrefix(path, separator),
	}
	return strings.Join(parts, separator)
}

func expToPointer(e jp.Expr) (string, error) {
	var buf []byte
	for _, f := range e {
		v, err := fragToPointer(f)
		if err != nil {
			return "", err
		}
		if v != nil {
			buf = append(buf, '/')
		}

		buf = append(buf, v...)
	}

	return string(buf), nil
}

func fragToPointer(f jp.Frag) ([]byte, error) {
	switch v := f.(type) {
	case jp.Root:
		return rootToPointer()
	case jp.Nth:
		return nthToPointer(v)
	case jp.Child:
		return toPointer(v)
	default:
		return nil, errors.New("There is no conversion implemented for " + reflect.TypeOf(v).Name())
	}
}

func rootToPointer() ([]byte, error) {
	return nil, nil
}

func nthToPointer(f jp.Nth) ([]byte, error) {
	var buf []byte
	i := int(f)
	if i < 0 {
		buf = append(buf, '-')
		i = -i
	}
	num := [20]byte{}
	cnt := 0
	for ; i != 0; cnt++ {
		num[cnt] = byte(i%10) + '0'
		i /= 10
	}
	if 0 < cnt {
		cnt--
		for ; 0 <= cnt; cnt-- {
			buf = append(buf, num[cnt])
		}
	} else {
		buf = append(buf, '0')
	}
	return buf, nil
}

func toPointer(f jp.Child) ([]byte, error) {
	var buf []byte

	// JSONPointer escaping https://datatracker.ietf.org/doc/html/rfc6901#section-3
	for _, b := range []byte(string(f)) {
		switch b {
		case '~':
			buf = append(buf, "~0"...)
		case '/':
			buf = append(buf, "~1"...)
		default:
			buf = append(buf, b)
		}
	}

	return buf, nil
}
