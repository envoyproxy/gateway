Added an opt-in endpoint fast path (`EndpointFastPath` runtime flag): EndpointSlice
updates are propagated to Envoy as EDS-only pushes using cached per-cluster context
from the last successful translation, so endpoint freshness no longer waits for a
full translation. Endpoint changes that affect configuration (e.g. a backend's
endpoint count crossing zero, or clusters affected by EnvoyPatchPolicy or extension
server hooks) still take the full path, which remains authoritative.
