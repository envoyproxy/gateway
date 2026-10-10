Fixed `healthCheck.passive.maxEjectionPercent: 0` being dropped from the generated Envoy cluster, which made Envoy fall back to its 10% default instead of prohibiting ejection.
