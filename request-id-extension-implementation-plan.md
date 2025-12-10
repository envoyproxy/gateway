# Implementation Plan: Request ID Extension Configuration in Envoy Gateway

## Overview

Add support for configuring Envoy's `request_id_extension` in the HTTP Connection Manager to enable custom request ID generation and trace ID behavior. This will allow users to control how trace IDs are generated, including deriving them from custom headers like Cloudflare's CF-Ray.

## Current Limitation

Envoy Gateway currently does NOT expose the `request_id_extension` field from Envoy's HCM. The trace ID is hardcoded to use Envoy's default UUID request ID extension with default parameters.

**Code reference:** `internal/xds/translator/listener.go:332` (buildHCMTracing is called but doesn't configure request_id_extension)

## Implementation Steps

### 1. API Changes - Add Request ID Configuration Types

**File:** `api/v1alpha1/envoyproxy_tracing_types.go`

Add new types to configure request ID extension:

```go
// RequestIDExtension defines the configuration for request ID generation.
// +kubebuilder:validation:XValidation:message="only one of uuid or custom can be set",rule="(has(self.uuid) && !has(self.custom)) || (!has(self.uuid) && has(self.custom))"
type RequestIDExtension struct {
	// UUID configures the default UUID-based request ID extension.
	// This is the default if no request ID extension is specified.
	// +optional
	UUID *UUIDRequestIDConfig `json:"uuid,omitempty"`
	
	// Custom allows specifying a custom request ID extension configuration.
	// This can be used to implement custom trace ID generation logic.
	// +optional
	Custom *CustomRequestIDExtension `json:"custom,omitempty"`
}

// UUIDRequestIDConfig configures the UUID request ID extension.
// This corresponds to envoy.extensions.request_id.uuid.v3.UuidRequestIdConfig
type UUIDRequestIDConfig struct {
	// PackTraceReason controls whether the UUID is modified to include
	// the trace sampling decision in the 14th nibble.
	// When true (default), the UUID nibble will be set to:
	// - '9': Sampled
	// - 'a': Force traced (server-side override)
	// - 'b': Force traced (client-side request ID joining)
	// When false, the UUID is not modified.
	// 
	// Disabling this may be necessary if you need to preserve externally
	// generated request IDs or implement custom trace ID derivation.
	// 
	// +optional
	// +kubebuilder:default=true
	PackTraceReason *bool `json:"packTraceReason,omitempty"`
	
	// UseRequestIDForTraceSampling controls whether to use the x-request-id
	// header for trace sampling decisions.
	// 
	// +optional
	// +kubebuilder:default=true
	UseRequestIDForTraceSampling *bool `json:"useRequestIdForTraceSampling,omitempty"`
}

// CustomRequestIDExtension allows specifying a custom request ID extension.
type CustomRequestIDExtension struct {
	// Name is the name of the custom request ID extension.
	// Examples: "envoy.request_id.custom_header"
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	
	// TypedConfig contains the configuration for the custom extension.
	// This should be a valid Envoy extension configuration in JSON format.
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	TypedConfig apiextensionsv1.JSON `json:"typedConfig"`
}
```

**Update ProxyTracing struct:**

```go
type ProxyTracing struct {
	// ... existing fields ...
	
	// RequestIDExtension configures how request IDs are generated.
	// If not specified, Envoy's default UUID request ID extension is used.
	// +optional
	RequestIDExtension *RequestIDExtension `json:"requestIdExtension,omitempty"`
}
```

### 2. Translator Changes - Wire Through Configuration

**File:** `internal/xds/translator/listener.go`

Modify the HCM builder to configure request_id_extension:

```go
// Around line 332, after buildHCMTracing
hcmTracing, err := buildHCMTracing(tracing)
if err != nil {
	return err
}

// Add new function call
requestIDExtension, err := buildRequestIDExtension(tracing)
if err != nil {
	return err
}

// Later when building HCM (around line 370+)
mgr := &hcmv3.HttpConnectionManager{
	// ... existing fields ...
	Tracing: hcmTracing,
	RequestIdExtension: requestIDExtension,  // Add this
	// ... rest of fields ...
}
```

**File:** `internal/xds/translator/tracing.go`

Add new function to build request ID extension config:

```go
import (
	uuidv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/request_id/uuid/v3"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func buildRequestIDExtension(tracing *ir.Tracing) (*corev3.TypedExtensionConfig, error) {
	if tracing == nil || tracing.RequestIDExtension == nil {
		// Return default UUID config
		return buildDefaultUUIDRequestIDExtension()
	}

	ext := tracing.RequestIDExtension
	
	if ext.UUID != nil {
		return buildUUIDRequestIDExtension(ext.UUID)
	}
	
	if ext.Custom != nil {
		return buildCustomRequestIDExtension(ext.Custom)
	}
	
	return buildDefaultUUIDRequestIDExtension()
}

func buildDefaultUUIDRequestIDExtension() (*corev3.TypedExtensionConfig, error) {
	config := &uuidv3.UuidRequestIdConfig{
		PackTraceReason:              wrapperspb.Bool(true),
		UseRequestIdForTraceSampling: wrapperspb.Bool(true),
	}
	
	anyConfig, err := proto.ToAnyWithValidation(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal UUID request ID config: %w", err)
	}
	
	return &corev3.TypedExtensionConfig{
		Name:        "envoy.request_id.uuid",
		TypedConfig: anyConfig,
	}, nil
}

func buildUUIDRequestIDExtension(cfg *egv1a1.UUIDRequestIDConfig) (*corev3.TypedExtensionConfig, error) {
	config := &uuidv3.UuidRequestIdConfig{
		PackTraceReason:              wrapperspb.Bool(ptr.Deref(cfg.PackTraceReason, true)),
		UseRequestIdForTraceSampling: wrapperspb.Bool(ptr.Deref(cfg.UseRequestIDForTraceSampling, true)),
	}
	
	anyConfig, err := proto.ToAnyWithValidation(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal UUID request ID config: %w", err)
	}
	
	return &corev3.TypedExtensionConfig{
		Name:        "envoy.request_id.uuid",
		TypedConfig: anyConfig,
	}, nil
}

func buildCustomRequestIDExtension(cfg *egv1a1.CustomRequestIDExtension) (*corev3.TypedExtensionConfig, error) {
	// Parse the JSON into an Any proto
	anyConfig := &anypb.Any{}
	if err := anyConfig.UnmarshalJSON(cfg.TypedConfig.Raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom request ID config: %w", err)
	}
	
	return &corev3.TypedExtensionConfig{
		Name:        cfg.Name,
		TypedConfig: anyConfig,
	}, nil
}
```

### 3. IR Changes - Add to Internal Representation

**File:** `internal/ir/xds.go`

Update the Tracing struct (around line 2472):

```go
type Tracing struct {
	ServiceName          string                      `json:"serviceName"`
	Authority            string                      `json:"authority,omitempty"`
	SamplingRate         float64                     `json:"samplingRate,omitempty"`
	CustomTags           map[string]egv1a1.CustomTag `json:"customTags,omitempty"`
	Destination          RouteDestination            `json:"destination,omitempty"`
	Traffic              *TrafficFeatures            `json:"traffic,omitempty"`
	Provider             egv1a1.TracingProvider      `json:"provider"`
	RequestIDExtension   *egv1a1.RequestIDExtension  `json:"requestIdExtension,omitempty"` // Add this
}
```

### 4. Gateway API Translator Changes

**File:** `internal/gatewayapi/listener.go`

Update the code that builds the IR Tracing struct (around line 757):

```go
irTracing := &ir.Tracing{
	// ... existing fields ...
	CustomTags:         tracing.CustomTags,
	RequestIDExtension: proxyTracing.RequestIDExtension, // Add this
}
```

### 5. Validation

**File:** `api/v1alpha1/validation/envoyproxy_validate.go`

Add validation for request ID extension:

```go
func validateRequestIDExtension(ext *egv1a1.RequestIDExtension) error {
	if ext == nil {
		return nil
	}
	
	// Validate that only one of uuid or custom is set (XValidation should catch this)
	if ext.UUID != nil && ext.Custom != nil {
		return errors.New("only one of uuid or custom can be specified")
	}
	
	// Validate custom extension
	if ext.Custom != nil {
		if ext.Custom.Name == "" {
			return errors.New("custom request ID extension name cannot be empty")
		}
		
		if len(ext.Custom.TypedConfig.Raw) == 0 {
			return errors.New("custom request ID extension typedConfig cannot be empty")
		}
		
		// Validate that typedConfig is valid JSON
		var js map[string]interface{}
		if err := json.Unmarshal(ext.Custom.TypedConfig.Raw, &js); err != nil {
			return fmt.Errorf("custom request ID extension typedConfig must be valid JSON: %w", err)
		}
	}
	
	return nil
}

// Add to validateProxyTracing function
func validateProxyTracing(tracing *egv1a1.ProxyTracing) error {
	// ... existing validation ...
	
	if err := validateRequestIDExtension(tracing.RequestIDExtension); err != nil {
		return err
	}
	
	return nil
}
```

**File:** `api/v1alpha1/validation/envoyproxy_validate_test.go`

Add test cases for request ID extension validation.

### 6. Test Coverage

#### Unit Tests

**File:** `internal/xds/translator/tracing_test.go`

Add tests for:
- Default UUID request ID extension generation
- Custom UUID config with packTraceReason=false
- Custom UUID config with useRequestIdForTraceSampling=false
- Custom request ID extension with valid JSON config
- Error cases (invalid JSON, etc.)

**File:** `api/v1alpha1/validation/envoyproxy_validate_test.go`

Add tests for:
- Valid UUID config
- Valid custom config
- Invalid: both UUID and custom set
- Invalid: empty custom name
- Invalid: empty custom typedConfig
- Invalid: malformed JSON in typedConfig

#### E2E Tests

**File:** `test/e2e/testdata/tracing-custom-request-id.yaml`

Create test manifest with custom request ID configuration:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: tracing-custom-request-id
  namespace: gateway-conformance-infra
spec:
  telemetry:
    tracing:
      provider:
        backendRefs:
          - name: otel-collector
            namespace: monitoring
            port: 4317
      requestIdExtension:
        uuid:
          packTraceReason: false
          useRequestIdForTraceSampling: true
```

**File:** `test/e2e/tests/tracing.go`

Add test case to verify request ID extension configuration is applied correctly.

### 7. Documentation

**File:** `site/content/en/latest/api/extension_types.md`

Document the new RequestIDExtension API types.

**File:** `site/content/en/latest/tasks/observability/tracing.md`

Add examples showing:
1. How to disable trace reason packing for externally-generated trace IDs
2. How to configure custom request ID extensions
3. Use case: integrating with Cloudflare Ray IDs

### 8. Generate Required Files

Run code generation:

```bash
make generate
```

This will update:
- `api/v1alpha1/zz_generated.deepcopy.go` - DeepCopy methods for new types
- `internal/ir/zz_generated.deepcopy.go` - DeepCopy methods for IR changes
- CRD manifests in `charts/gateway-helm/crds/`

## Implementation Notes

### Why This Approach?

1. **Start Simple**: Expose UUID config first (pack_trace_reason, use_request_id_for_trace_sampling)
2. **Future Extensibility**: Custom extension support allows advanced users to implement custom logic
3. **Safe Defaults**: If not specified, behavior remains unchanged (default UUID with trace reason packing)
4. **Validation**: Strong validation prevents misconfiguration

### Limitations

This implementation still **does not allow deriving trace IDs from arbitrary headers** like CF-Ray directly. It only exposes Envoy's built-in request ID extensions.

To derive trace IDs from CF-Ray, users would need to:

**Option A:** Create a custom Envoy request_id extension (C++ code, compile into Envoy)
**Option B:** Use OTel Collector transform processor (recommended - no code changes needed)
**Option C:** Implement a Wasm plugin for custom request ID generation

### Future Enhancements

1. **Built-in Header-Based Request ID Extension**
   - Add a new extension type that derives request IDs from headers
   - Implement in Go as an external processor
   - More complex, requires additional infrastructure

2. **Wasm Plugin Support**
   - Add configuration for loading Wasm plugins that implement custom request ID logic
   - Most flexible but requires users to write Wasm

3. **Direct CF-Ray Integration**
   - Add first-class support for Cloudflare Ray ID conversion
   - Built-in conversion function from Ray ID to trace ID
   - Easiest for users but very specific use case

## Example Usage

### Basic: Disable Trace Reason Packing

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
      requestIdExtension:
        uuid:
          packTraceReason: false
```

### Advanced: Custom Extension (Future)

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
      requestIdExtension:
        custom:
          name: "envoy.request_id.custom_header"
          typedConfig:
            "@type": "type.googleapis.com/envoy.extensions.request_id.custom_header.v3.CustomHeaderConfig"
            header_name: "cf-ray"
            conversion_algorithm: "deterministic_hash"
```

## Testing Strategy

1. **Unit Tests** - Verify correct Envoy config generation
2. **Validation Tests** - Ensure API validation catches errors
3. **Integration Tests** - Verify config is applied to running Envoy
4. **E2E Tests** - Verify tracing behavior with custom request ID config

## Timeline Estimate

- API changes: 2-4 hours
- Translator implementation: 4-6 hours
- Validation: 2-3 hours
- Unit tests: 3-4 hours
- E2E tests: 2-3 hours
- Documentation: 2-3 hours
- Code review iterations: 2-4 hours

**Total: 17-27 hours** (2-3 days of focused work)

## Breaking Changes

None. This is a purely additive change. Existing configurations will continue to work with default behavior.

## Alternative Approaches Considered

### 1. Only Expose PackTraceReason Boolean

**Pros:** Simpler API, covers most use cases
**Cons:** Less flexible, can't support custom extensions

### 2. Full Request ID Extension Config

**Pros:** Maximum flexibility
**Cons:** Complex API, harder to validate, security concerns with arbitrary extensions

### 3. CF-Ray Specific Field

**Pros:** Easy to use for Cloudflare users
**Cons:** Too specific, doesn't help other use cases, not extensible

**Chosen approach balances flexibility with usability.**

## Success Criteria

1. ✅ Users can disable trace reason packing in UUIDs
2. ✅ Users can configure UUID request ID extension parameters
3. ✅ Configuration is validated at admission time
4. ✅ Changes generate correct Envoy xDS configuration
5. ✅ E2E tests verify behavior
6. ✅ Documentation explains use cases and examples
7. ✅ No breaking changes to existing configurations

## Related Issues

- [Link to GitHub issue if one exists]
- Similar work in other projects (Istio, Contour, etc.)

## References

- [Envoy Request ID Documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#x-request-id)
- [UUID Request ID Extension Proto](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/request_id/uuid/v3/uuid.proto)
- [go-control-plane UUID Package](https://pkg.go.dev/github.com/envoyproxy/go-control-plane/envoy/extensions/request_id/uuid/v3)
