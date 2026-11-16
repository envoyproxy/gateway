Fixed endpoint changes of Services referenced only by a SecurityPolicy's OIDC provider backendRefs not being reconciled, which left Envoy calling a stale token endpoint IP.
