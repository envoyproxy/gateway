Added a `detailedMetric` field to global rate limit rules in BackendTrafficPolicy,
exposing the rate limit service's `detailed_metric` descriptor flag. When enabled,
the near-limit and over-limit metrics include the resolved descriptor value,
making them per-value rather than aggregated, at the cost of potentially high
metric cardinality for descriptors with many distinct values. The field is
rejected for local rate limits, and when unset the previous aggregated behavior is
preserved.
