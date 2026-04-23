# OpenTelemetry Example

This example demonstrates OpenTelemetry metrics, tracing, and access logging
with custom headers. It uses Envoy Gateway's standalone mode (no Kubernetes
required).

Note: This uses a fake token to demonstrate how headers are sent to the OTLP
collector. Production systems should use proper secret management.

## Architecture

```
┌──────────┐        ┌───────────────────┐        ┌─────────────┐
│   curl   │───────▶│       Envoy       │───────▶│ httpbingo.org │
└──────────┘ :10080 └───────────────────┘        └─────────────┘
                             │
                             │ OTLP/gRPC :4317
                             │ metrics + traces + access logs
                             │ Authorization: Bearer fake
                             ▼
                    ┌────────────────────────┐
                    │ AUTH_TOKEN=fake        │
                    │ otel-tui               │
                    └────────────────────────┘
```

## Running the Example

1. Start otel-tui with auth token validation:
   ```bash
   AUTH_TOKEN=fake otel-tui
   ```

2. From this directory, start Envoy Gateway:
   ```bash
   envoy-gateway server --config-path envoy-gateway.yaml
   ```

3. Send a request:
   ```bash
   curl localhost:10080/get
   ```

4. In otel-tui, you should see metrics, traces, and access logs. Traces and
   access logs share the same trace ID. If the Authorization header was missing
   or wrong, otel-tui would reject the OTLP export.

## Files

- [envoy-gateway.yaml](envoy-gateway.yaml) - EnvoyGateway configuration
- [resources/gateway.yaml](resources/gateway.yaml) - Gateway API resources
