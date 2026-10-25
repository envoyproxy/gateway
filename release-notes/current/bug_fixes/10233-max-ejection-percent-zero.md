Fixed `healthCheck.passive.maxEjectionPercent: 0` being dropped from the generated Envoy cluster; the explicitly configured 0% limit is now preserved instead of falling back to Envoy's 10% default.
