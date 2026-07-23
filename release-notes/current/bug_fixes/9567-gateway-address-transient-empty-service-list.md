Fixed a Gateway becoming permanently stuck at `Programmed=False` / `AddressNotAssigned` with its
status addresses cleared, after a transient empty result from the cached client when looking up the
Envoy LoadBalancer Service. Such an empty result (e.g. an informer cache that has not yet observed
the Service, or a relist window) was treated as "the Service no longer exists", wiping the Gateway's
status addresses. The lookup now confirms with an uncached read against the API server before
concluding the Service is gone.
