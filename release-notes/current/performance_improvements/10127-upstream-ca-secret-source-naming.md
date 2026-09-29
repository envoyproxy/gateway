Reduced xDS configuration size and Envoy data-plane memory used to store SDS secrets by emitting one SDS secret per CA source object instead of one per `BackendTLSPolicy` or `Backend`, so policies trusting the same ConfigMap, Secret or ClusterTrustBundle no longer each ship a copy of the same bundle.

Reduced Envoy Gateway control-plane memory usage by storing each shared upstream CA bundle once per gateway in the intermediate representation (IR), with route destinations referencing it by name. This avoids duplicating certificate bytes for every destination when the IR is deep-copied.
