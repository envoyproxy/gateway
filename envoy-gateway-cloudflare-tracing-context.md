# Envoy Gateway + Cloudflare Ray ID Tracing Context

## Problem Statement

We're introducing Envoy Gateway into our architecture and need to maintain Cloudflare as the parent trace in our distributed tracing setup.

### Current Architecture
```
Cloudflare → k8s LB → Backend
```

The backend calculates a parent trace ID based on the Cloudflare Ray ID. We process Cloudflare logs, send the span, and our trace collector correctly makes the Cloudflare span the parent of the backend spans.

### Target Architecture
```
Cloudflare → Envoy Gateway → Backend
```

We need to preserve the behavior where Cloudflare is the root/parent trace, with Envoy Gateway and backend spans as children underneath.

## Envoy Gateway Tracing Capabilities

### What's Configurable

Envoy Gateway exposes OpenTelemetry tracing configuration through the `EnvoyProxy` CRD:

**Location in codebase:**
- API types: `api/v1alpha1/envoyproxy_tracing_types.go`
- Translator: `internal/xds/translator/tracing.go:35` (buildHCMTracing)
- IR types: `internal/ir/xds.go:2472`

**Available configuration options:**
1. **CustomTags** - Add custom tags to spans (Literal, Environment, RequestHeader types)
2. **SamplingRate/SamplingFraction** - Control trace sampling
3. **ServiceName** - Customize service name in traces
4. **Provider** - Configure OpenTelemetry, Zipkin, or Datadog backends via `backendRefs`

**Example configuration:**
```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
spec:
  telemetry:
    tracing:
      provider:
        backendRefs:
          - name: otel-collector
            namespace: monitoring
            port: 4317
      customTags:
        "cf-ray":
          type: RequestHeader
          requestHeader:
            name: cf-ray
            defaultValue: "-"
```

### What's NOT Configurable

Envoy Gateway does NOT expose:
- `request_id_extension` field from Envoy's HTTP Connection Manager
- Direct control over trace ID generation
- Ability to calculate trace IDs from headers

The trace ID is managed by Envoy's core tracing infrastructure and uses W3C traceparent header propagation for OpenTelemetry.

**To modify trace ID behavior would require:**
1. Extending the API (`api/v1alpha1/envoyproxy_tracing_types.go`)
2. Updating the translator (`internal/xds/translator/tracing.go` and `listener.go`)
3. Adding validation (`api/v1alpha1/validation/envoyproxy_validate.go`)

This would be a code change to Envoy Gateway itself, not a configuration option.

## Solution: OpenTelemetry Collector Transform Processor

### Architecture Overview

Envoy Gateway does NOT include an OTel Collector - it only generates spans and sends them to an external endpoint.

**Required architecture:**
```
Envoy Gateway → OTel Collector (transform processor) → Refinery → Trace Backend
                  ↑
              You deploy this separately
```

### Why OTel Collector Transform Processor

The OTel Collector's transform processor **can modify trace_id and parent_span_id** based on span attributes.

**Capabilities:**
- Read attributes from spans (like CF-Ray header captured as custom tag)
- Calculate deterministic trace_id from CF-Ray
- Set parent_span_id to link spans to Cloudflare parent trace
- Uses OTTL (OpenTelemetry Transformation Language)

**Warning:** OTel docs note "modifying these fields could lead to orphaned spans or logs" - but this is acceptable for our use case since we're intentionally re-parenting spans.

**Reference:** [OTel Collector Transform Processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/transformprocessor/README.md)

### Implementation Steps

1. **Configure Envoy Gateway** to capture CF-Ray as custom tag (shown above)

2. **Deploy OTel Collector** in your cluster with transform processor

3. **Configure Transform Processor** to:
   - Extract CF-Ray from span attributes
   - Calculate trace_id deterministically from CF-Ray
   - Set parent_span_id appropriately
   - Forward modified spans to Refinery

4. **Point Envoy Gateway** to OTel Collector via `backendRefs`

5. **Configure OTel Collector** to forward to Refinery

### Why Not Refinery?

Refinery **cannot** modify trace/span IDs. It only:
- Reads `TraceNames` and `ParentNames` config for routing/sampling decisions
- Makes keep/drop decisions on traces based on sampling rules
- Does NOT rewrite trace context

**Reference:** [Refinery Configuration Docs](https://docs.honeycomb.io/manage-data-volume/sample/honeycomb-refinery/configure/)

## Key Findings from Research

### Envoy Tracing Headers

- Envoy uses `x-request-id` for internal request correlation (separate from trace ID)
- OpenTelemetry tracer uses W3C `traceparent` header for trace context propagation
- Trace ID generation happens via Envoy's UUID request ID extension by default
- Custom request ID extensions exist but aren't exposed by Envoy Gateway

### Cloudflare Ray ID

- Cloudflare Ray ID is added as `cf-ray` request header
- Format: `abc123-SJC` (unique ID + datacenter code)
- Can be used to correlate Cloudflare logs with application traces
- Needs deterministic conversion to OpenTelemetry trace_id format (128-bit hex)

### OpenTelemetry Trace Context

- `trace_id`: 128-bit hex identifier for the entire trace
- `span_id`: 64-bit hex identifier for a single span
- `parent_span_id`: Links child spans to parent spans
- Root spans have no parent_span_id (all zeros or absent)

## Additional Context

### Current Backend Implementation

Your backend currently:
1. Receives CF-Ray header from requests
2. Calculates parent trace ID deterministically from CF-Ray
3. Creates spans with that parent trace ID
4. Posts Cloudflare log-based spans to trace collector
5. Trace collector correctly links everything together

### Challenge with Envoy Gateway

Envoy Gateway will:
1. Generate its own trace_id (not based on CF-Ray)
2. Propagate that trace_id to backend via `traceparent` header
3. Backend receives both CF-Ray and Envoy's trace context
4. Need to ensure Envoy's spans and backend spans share same trace_id derived from CF-Ray

### Solution Flow

```
1. Cloudflare adds cf-ray: abc123-SJC
2. Envoy Gateway:
   - Captures cf-ray as span attribute (custom tag)
   - Generates span with temporary trace_id
   - Sends to OTel Collector
3. OTel Collector Transform Processor:
   - Reads cf-ray attribute: "abc123-SJC"
   - Calculates trace_id: hash("abc123-SJC") → "a1b2c3d4..."
   - Sets parent_span_id: cloudflare_span_id
   - Forwards modified span
4. OTel Collector sends to Refinery
5. Refinery applies sampling, forwards to backend
6. Backend receives modified traces with correct parent
```

## Files to Reference

- `test/e2e/testdata/tracing-otel.yaml` - Example Envoy Gateway tracing config
- `api/v1alpha1/envoyproxy_tracing_types.go` - Tracing type definitions
- `internal/xds/translator/tracing.go` - Tracing configuration translator
- `internal/ir/xds.go:2472` - Internal representation of tracing config

## Questions to Answer in Implementation

1. What algorithm should convert CF-Ray to trace_id deterministically?
2. Should we generate a static parent_span_id for Cloudflare or calculate it?
3. How to handle requests without CF-Ray header (local testing)?
4. What happens if OTel Collector is down - fallback behavior?
5. Performance impact of transform processor on trace throughput?

## Next Steps

Need to create:
1. OTel Collector deployment manifest for k8s
2. OTel Collector configuration with transform processor
3. OTTL statements to perform trace_id/parent_span_id transformation
4. Updated Envoy Gateway configuration with customTags for cf-ray
5. Testing strategy to validate parent-child relationships
