The flat fields on `ClientTrafficPolicy.spec.http1` (`enableTrailers`, `preserveHeaderCase`, `http10`, `disableSafeMaxConnectionDuration`, `ignoredUpgradeTypes`) are deprecated. Use the equivalent fields under `spec.http1.client` instead (e.g. `spec.http1.client.preserveHeaderCase`). When `spec.http1.client` is set, all flat `spec.http1` fields are ignored. The flat fields will be removed in a future API version.

For upstream HTTP/1 settings, use the new `spec.http1` field on `BackendTrafficPolicy` instead.
