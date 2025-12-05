# OTLP Access Log Headers Example

This example demonstrates how to configure custom headers for OpenTelemetry
Protocol (OTLP) access log exports. Custom headers are useful for routing or
authentication when exporting logs to OTLP collectors.

While production systems discourage Authorization headers defined in ENV, this
example uses a static token for simplicity. This also uses Envoy Gateway's host
(a.k.a. standalone mode), which does not require a Kubernetes cluster.

## Architecture

```
┌──────────┐        ┌───────────────────┐        ┌─────────────┐
│   curl   │───────▶│       Envoy       │───────▶│ httpbin.org │
└──────────┘ :10080 └───────────────────┘        └─────────────┘
                             │
                             │ OTLP/gRPC :4317
                             │ Authorization: Bearer fake
                             ▼
                    ┌────────────────────────┐
                    │ AUTH_TOKEN=fake        │
                    │ otel-tui               │
                    └────────────────────────┘
```

Envoy Gateway is the control plane that configures Envoy. Envoy handles all
traffic and exports access logs directly to the OTLP collector.

## Flow

1. **curl** sends an HTTP request to Envoy on port 10080 with `Host: 127.0.0.1.nip.io`
2. **Envoy** matches the HTTPRoute and proxies the request to httpbin.org
3. **Envoy** generates an access log for the request
4. **Envoy** exports logs via OTLP/gRPC to localhost:4317 with `Authorization: Bearer fake` header
5. **otel-tui** validates the Authorization header and displays the logs

## Running the Example

1. Start otel-tui with auth token validation:
   ```bash
   AUTH_TOKEN=fake otel-tui
   ```

2. From this directory, start Envoy Gateway:
   ```bash
   envoy-gateway server --config-path envoy-gateway.yaml
   ```

3. Send traffic through the gateway:
   ```bash
   curl -H "Host: 127.0.0.1.nip.io" localhost:10080/get
   ```

4. Check otel-tui - you should see logs. If the Authorization header was
   missing or wrong, otel-tui would reject the OTLP export.

## Files

- [envoy-gateway.yaml](envoy-gateway.yaml) - EnvoyGateway configuration for standalone mode
- [resources/gateway.yaml](resources/gateway.yaml) - Gateway API resources including Access Logging and Otel
