Fixed local rate limiting for Distinct client selectors to retain up to 10,000
per-value token buckets per wildcard descriptor by applying the cache limit to
each route's filter configuration instead of relying on Envoy's default of 20.
