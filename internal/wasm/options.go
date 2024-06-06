// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.
//
// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wasm

import (
	"time"

	"github.com/envoyproxy/gateway/internal/utils/sets"
)

const (
	DefaultPurgeInterval         = 1 * time.Hour
	DefaultModuleExpiry          = 24 * time.Hour
	DefaultHTTPRequestTimeout    = 15 * time.Second
	DefaultHTTPRequestMaxRetries = 5
	DefaultPullTimeout           = 5 * time.Minute
)

// Options contains configurations to create a Cache instance.
type Options struct {
	PurgeInterval         time.Duration
	ModuleExpiry          time.Duration
	InsecureRegistries    sets.String
	HTTPRequestTimeout    time.Duration
	HTTPRequestMaxRetries int
}

func defaultOptions() Options {
	return Options{
		PurgeInterval:         DefaultPurgeInterval,
		ModuleExpiry:          DefaultModuleExpiry,
		InsecureRegistries:    sets.New[string](),
		HTTPRequestTimeout:    DefaultHTTPRequestTimeout,
		HTTPRequestMaxRetries: DefaultHTTPRequestMaxRetries,
	}
}

type PullPolicy int32

const (
	Unspecified  PullPolicy = 0
	IfNotPresent PullPolicy = 1
	Always       PullPolicy = 2
)

var PullPolicyValues = map[string]PullPolicy{
	"UNSPECIFIED_POLICY": 0,
	"IfNotPresent":       1,
	"Always":             2,
}

// GetOptions is a struct for providing options to Get method of Cache.
type GetOptions struct {
	Checksum        string
	ResourceName    string
	ResourceVersion string
	RequestTimeout  time.Duration
	PullSecret      []byte
	PullPolicy      PullPolicy
}

