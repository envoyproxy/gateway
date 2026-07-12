Hardened the gateway-helm chart: the Envoy Gateway controller container now runs with a read-only root filesystem, and the certgen Job runs with a restricted pod-level securityContext by default.
