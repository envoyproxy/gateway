// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package extensionserver

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

type htpasswd struct {
	Users map[string]string
}

func NewHtpasswd() htpasswd {
	return htpasswd{
		Users: map[string]string{},
	}
}

func (h *htpasswd) AddUser(user, password string) {
	s := sha1.New()
	io.WriteString(s, password)
	h.Users[user] = fmt.Sprintf("{SHA}%s", base64.StdEncoding.EncodeToString(s.Sum(nil)))
}

func (h *htpasswd) String() string {
	var b strings.Builder
	for user, password := range h.Users {
		b.WriteString(fmt.Sprintf("%s:%s\n", user, password))
	}
	return b.String()
}
