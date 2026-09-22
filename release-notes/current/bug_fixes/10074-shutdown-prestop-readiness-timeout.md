Fixed the Envoy proxy preStop readiness request failing after a graceful drain because the shutdown manager's HTTP write deadline was shorter than the configured readiness timeout.
