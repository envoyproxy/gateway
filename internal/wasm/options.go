// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

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

	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	DefaultPurgeInterval         = 1 * time.Hour
	DefaultModuleExpiry          = 24 * time.Hour
	DefaultHTTPRequestTimeout    = 15 * time.Second
	DefaultHTTPRequestMaxRetries = 5
	DefaultPullTimeout           = 5 * time.Minute
	DefaultMaxCacheSize          = 1024 * 1024 * 1024 // 1GB
)

// CacheOptions contains configurations to create a Cache instance.
type CacheOptions struct {
	PurgeInterval time.Duration
	ModuleExpiry  time.Duration
	// InsecureRegistries is a set of registries that are allowed to be accessed without TLS.
	InsecureRegistries    sets.Set[string]
	HTTPRequestTimeout    time.Duration
	HTTPRequestMaxRetries int
	MaxCacheSize          int
	CacheDir              string
}

// allowInsecure returns true if the host is allowed to be accessed without TLS.
func (o *CacheOptions) allowInsecure(host string) bool {
	return o.InsecureRegistries.Has(host) || o.InsecureRegistries.Has("*")
}

func (o *CacheOptions) sanitize() CacheOptions {
	ret := defaultCacheOptions()
	if o.InsecureRegistries != nil {
		ret.InsecureRegistries = o.InsecureRegistries
	}
	if o.PurgeInterval != 0 {
		ret.PurgeInterval = o.PurgeInterval
	}
	if o.ModuleExpiry != 0 {
		ret.ModuleExpiry = o.ModuleExpiry
	}
	if o.HTTPRequestTimeout != 0 {
		ret.HTTPRequestTimeout = o.HTTPRequestTimeout
	}
	if o.HTTPRequestMaxRetries != 0 {
		ret.HTTPRequestMaxRetries = o.HTTPRequestMaxRetries
	}
	if o.MaxCacheSize != 0 {
		ret.MaxCacheSize = o.MaxCacheSize
	}
	if o.CacheDir != "" {
		ret.CacheDir = o.CacheDir
	}

	return ret
}

func defaultCacheOptions() CacheOptions {
	return CacheOptions{
		PurgeInterval:         DefaultPurgeInterval,
		ModuleExpiry:          DefaultModuleExpiry,
		InsecureRegistries:    sets.New[string](),
		HTTPRequestTimeout:    DefaultHTTPRequestTimeout,
		HTTPRequestMaxRetries: DefaultHTTPRequestMaxRetries,
		MaxCacheSize:          DefaultMaxCacheSize,
	}
}

type PullPolicy int32

const (
	Unspecified  PullPolicy = 0
	IfNotPresent PullPolicy = 1
	Always       PullPolicy = 2
)

// GetOptions is a struct for providing options to Get method of Cache.
type GetOptions struct {
	Checksum        string
	ResourceName    string
	ResourceVersion string
	RequestTimeout  time.Duration
	PullSecret      []byte
	PullPolicy      PullPolicy
}
