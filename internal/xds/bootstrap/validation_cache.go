// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package bootstrap

import (
	"crypto/sha256"
	"encoding/json"

	"k8s.io/utils/lru"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// ValidationCache caches the result of Validate by the content of the bootstrap.
type ValidationCache struct {
	cache    *lru.Cache
	validate func(*egv1a1.ProxyBootstrap) error
}

type validationResult struct{ err error }

// NewValidationCache returns cache that holds up to size distinct bootstraps,
// identical bootstraps share one entry. A size of zero means no limit.
func NewValidationCache(size int) *ValidationCache {
	return &ValidationCache{
		cache:    lru.New(size),
		validate: Validate,
	}
}

// Validate returns the same result as the package level Validate.
func (c *ValidationCache) Validate(cfg *egv1a1.ProxyBootstrap) error {
	if cfg == nil {
		return nil
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return c.validate(cfg)
	}

	key := sha256.Sum256(data)
	if v, ok := c.cache.Get(key); ok {
		return v.(validationResult).err
	}

	err = c.validate(cfg)
	c.cache.Add(key, validationResult{err: err})
	return err
}

// Validator validates a bootstrap on every call, without caching.
type Validator struct{}

func (Validator) Validate(cfg *egv1a1.ProxyBootstrap) error {
	return Validate(cfg)
}
