---
title: "Private Link Identifiers"
---

When traffic reaches Envoy Gateway through an AWS VPC Endpoint or Azure Private Endpoint, the cloud
provider injects the endpoint's identifier into the Proxy Protocol v2 (PPv2) header as a custom
Type-Length-Value (TLV) field. Extracting this identifier allows backends to distinguish which
Private Link connection originated a request; useful for per-tenant routing, audit logging, or
access control.

This task shows how to configure Envoy Gateway to parse the cloud-injected TLV and forward the
identifier as a plain HTTP request header.

## Background

Both AWS and Azure use PPv2 custom TLVs to carry Private Link identifiers:

| Cloud | TLV type | Value format                                   | Extracted as      |
|-------|----------|------------------------------------------------|-------------------|
| AWS   | `0xEA`   | 1 sub-type byte + VPC Endpoint ID (ASCII)      | `x-aws-vpce-id`   |
| Azure | `0xEE`   | 1 sub-type byte + link identifier (uint32, LE) | `x-azure-link-id` |

This behaviour is documented in the [AWS NLB target group attributes][aws-pp-spec] and the
[Azure Private Link service overview][azure-pp-spec].

The sub-type byte (always `0x01`) is a vendor prefix that must be stripped before forwarding the
value. Because the extraction logic is not yet expressible through standard Envoy Gateway APIs, both
cases use an [EnvoyPatchPolicy][] with two JSON patches: one to reconfigure the built-in
`proxy_protocol` listener filter so it captures the TLV, and one to prepend a Lua HTTP filter that
reads the captured value and adds the request header.

## Prerequisites

- Envoy Gateway must be deployed with `EnvoyPatchPolicy` enabled. See [Envoy Patch Policy][].
- The `XDSNameSchemeV2` [runtime flag][xds-name-scheme-v2] must be enabled — the JSON patch paths
  used in this task rely on the v2 xDS name scheme. Add the flag to your `EnvoyGateway`
  configuration:

  ```yaml
  runtimeFlags:
    enabled:
    - XDSNameSchemeV2
  ```

- The Gateway's [ClientTrafficPolicy][] must enable Proxy Protocol so that Envoy parses the PPv2
  header at all:

  ```yaml
  proxyProtocol:
    optional: false
  ```

- Downstream infrastructure (NLB on AWS, Internal Load Balancer on Azure) must be configured to
  forward PPv2 headers to the Envoy pods.

These examples target an HTTPS listener (`tcp-443`) which places the HTTP
connection manager inside `filter_chains`; the `jsonPath: "$.filter_chains[*]"` selector patches
every chain, which handles gateways with multiple TLS hostnames.

For a plain HTTP listener (`tcp-80`), the HCM is placed under `defaultFilterChain` instead, which is not part of
`filter_chains`, so the JSONPath selector matches nothing and the policy would be rejected. If your
Gateway listener is plain HTTP, replace the listener name with `tcp-80`, remove the `jsonPath`
field, and set `path` to `/defaultFilterChain/filters/0/typed_config/http_filters/0`.

## AWS — extracting the VPC Endpoint ID

AWS encodes the VPC Endpoint ID (e.g. `vpce-0123456789abcdef0`) in TLV `0xEA`. Like the Azure
case below, the TLV bytes are stored via `DYNAMIC_METADATA` and read through
`connectionStreamInfo():dynamicTypedMetadata()` in the Lua filter. Because the VPCE ID is ASCII
the sanitization concern does not apply, but the same approach is used for consistency.

Apply the following `EnvoyPatchPolicy` targeting your Gateway:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyPatchPolicy
metadata:
  name: extract-aws-vpce-id
  namespace: default
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: my-gateway
  type: JSONPatch
  jsonPatches:
    # Replace the plain proxy_protocol listener filter with one that captures
    # TLV 0xEA into DYNAMIC_METADATA (connection-level typed_filter_metadata).
    - type: "type.googleapis.com/envoy.config.listener.v3.Listener"
      name: "tcp-443"
      operation:
        op: replace
        path: "/listener_filters/0"
        value:
          name: envoy.filters.listener.proxy_protocol
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.listener.proxy_protocol.v3.ProxyProtocol
            rules:
              - tlv_type: 0xEA
                on_tlv_present:
                  metadata_namespace: "envoy.filters.listener.proxy_protocol"
                  key: "aws_vpce_id"
            pass_through_tlvs:
              tlv_type: [0xEA]
    # Prepend a Lua HTTP filter that reads the VPCE ID from connection-level
    # DYNAMIC_METADATA and sets the x-aws-vpce-id request header.
    - type: "type.googleapis.com/envoy.config.listener.v3.Listener"
      name: "tcp-443"
      operation:
        op: add
        jsonPath: "$.filter_chains[*]"
        path: "/filters/0/typed_config/http_filters/0"
        value:
          name: envoy.filters.http.lua
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
            default_source_code:
              inline_string: |
                function envoy_on_request(request_handle)
                  local meta = request_handle:connectionStreamInfo():dynamicTypedMetadata(
                    "envoy.filters.listener.proxy_protocol")
                  if not meta or not meta.typed_metadata then return end
                  local v = meta.typed_metadata["aws_vpce_id"]
                  if v and #v > 1 then
                    request_handle:headers():replace("x-aws-vpce-id", v:sub(2))
                  end
                end
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}

Save and apply the following resource to your cluster:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyPatchPolicy
metadata:
  name: extract-aws-vpce-id
  namespace: default
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: my-gateway
  type: JSONPatch
  jsonPatches:
    - type: "type.googleapis.com/envoy.config.listener.v3.Listener"
      name: "tcp-443"
      operation:
        op: replace
        path: "/listener_filters/0"
        value:
          name: envoy.filters.listener.proxy_protocol
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.listener.proxy_protocol.v3.ProxyProtocol
            rules:
              - tlv_type: 0xEA
                on_tlv_present:
                  metadata_namespace: "envoy.filters.listener.proxy_protocol"
                  key: "aws_vpce_id"
            pass_through_tlvs:
              tlv_type: [0xEA]
    - type: "type.googleapis.com/envoy.config.listener.v3.Listener"
      name: "tcp-443"
      operation:
        op: add
        jsonPath: "$.filter_chains[*]"
        path: "/filters/0/typed_config/http_filters/0"
        value:
          name: envoy.filters.http.lua
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
            default_source_code:
              inline_string: |
                function envoy_on_request(request_handle)
                  local meta = request_handle:connectionStreamInfo():dynamicTypedMetadata(
                    "envoy.filters.listener.proxy_protocol")
                  if not meta or not meta.typed_metadata then return end
                  local v = meta.typed_metadata["aws_vpce_id"]
                  if v and #v > 1 then
                    request_handle:headers():replace("x-aws-vpce-id", v:sub(2))
                  end
                end
```

{{% /tab %}}
{{< /tabpane >}}

Verify the policy was accepted:

```shell
kubectl get envoypatchpolicy/extract-aws-vpce-id -o yaml
```

Once traffic flows through a VPC Endpoint, the backend will receive an `x-aws-vpce-id` header whose
value matches the VPC Endpoint ID (e.g. `vpce-0123456789abcdef0`).

## Azure — extracting the Private Link identifier

Azure encodes the `linkIdentifier` (a 32-bit unsigned integer assigned per Private Endpoint
connection) in TLV `0xEE`. The raw bytes look like:

```plaintext
Type:   0xEE (238)
Length: 5
Value:  0x01 <4 bytes, little-endian uint32>
        ^^^^ sub-type byte (always 0x01)
```

### Why DYNAMIC_METADATA instead of FILTER_STATE

Envoy's proxy protocol implementation sanitizes TLV bytes before storing them in FILTER_STATE by
replacing bytes ≥ `0x80` with `0x21`. Because the `linkIdentifier` is a raw integer it frequently
contains high bytes (e.g. a link ID of `0x0047b001` encodes as `01 01 b0 2c 00 47`), which would
be corrupted by sanitization.

The workaround is to omit `tlv_location`, which defaults to `DYNAMIC_METADATA`. This path stores
the **unsanitized** bytes in `typed_filter_metadata` on the connection's `StreamInfo`. The Lua
filter then reads them via `connectionStreamInfo():dynamicTypedMetadata()`, which accesses
connection-level metadata rather than the per-request metadata read by `streamInfo()`.

This requires Envoy ≥ v1.32 (the `connectionStreamInfo()` API was added in that release).

Apply the following `EnvoyPatchPolicy` targeting your Gateway:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyPatchPolicy
metadata:
  name: extract-azure-link-id
  namespace: default
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: my-gateway
  type: JSONPatch
  jsonPatches:
    # Replace the plain proxy_protocol listener filter with one that captures TLV 0xEE
    # into DYNAMIC_METADATA. Omitting tlv_location keeps raw (unsanitized) bytes, which
    # is required because the link ID may contain bytes >= 0x80.
    - type: "type.googleapis.com/envoy.config.listener.v3.Listener"
      name: "tcp-443"
      operation:
        op: replace
        path: "/listener_filters/0"
        value:
          name: envoy.filters.listener.proxy_protocol
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.listener.proxy_protocol.v3.ProxyProtocol
            rules:
              - tlv_type: 0xEE
                on_tlv_present:
                  metadata_namespace: "envoy.filters.listener.proxy_protocol"
                  key: "azure_link_id"
            pass_through_tlvs:
              tlv_type: [0xEE]
    # Prepend a Lua HTTP filter that reads the 5-byte raw value from connection-level
    # typed_filter_metadata, skips the sub-type byte, decodes the remaining 4 bytes as
    # a little-endian uint32, and sets the x-azure-link-id request header.
    - type: "type.googleapis.com/envoy.config.listener.v3.Listener"
      name: "tcp-443"
      operation:
        op: add
        jsonPath: "$.filter_chains[*]"
        path: "/filters/0/typed_config/http_filters/0"
        value:
          name: envoy.filters.http.lua
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
            default_source_code:
              inline_string: |
                function envoy_on_request(request_handle)
                  local meta = request_handle:connectionStreamInfo():dynamicTypedMetadata(
                    "envoy.filters.listener.proxy_protocol")
                  if not meta or not meta.typed_metadata then return end
                  local v = meta.typed_metadata["azure_link_id"]
                  if v and #v == 5 then
                    local b1, b2, b3, b4 = v:byte(2), v:byte(3), v:byte(4), v:byte(5)
                    request_handle:headers():replace(
                      "x-azure-link-id",
                      tostring(b1 + b2*256 + b3*65536 + b4*16777216))
                  end
                end
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}

Save and apply the following resource to your cluster:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyPatchPolicy
metadata:
  name: extract-azure-link-id
  namespace: default
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: my-gateway
  type: JSONPatch
  jsonPatches:
    - type: "type.googleapis.com/envoy.config.listener.v3.Listener"
      name: "tcp-443"
      operation:
        op: replace
        path: "/listener_filters/0"
        value:
          name: envoy.filters.listener.proxy_protocol
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.listener.proxy_protocol.v3.ProxyProtocol
            rules:
              - tlv_type: 0xEE
                on_tlv_present:
                  metadata_namespace: "envoy.filters.listener.proxy_protocol"
                  key: "azure_link_id"
            pass_through_tlvs:
              tlv_type: [0xEE]
    - type: "type.googleapis.com/envoy.config.listener.v3.Listener"
      name: "tcp-443"
      operation:
        op: add
        jsonPath: "$.filter_chains[*]"
        path: "/filters/0/typed_config/http_filters/0"
        value:
          name: envoy.filters.http.lua
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
            default_source_code:
              inline_string: |
                function envoy_on_request(request_handle)
                  local meta = request_handle:connectionStreamInfo():dynamicTypedMetadata(
                    "envoy.filters.listener.proxy_protocol")
                  if not meta or not meta.typed_metadata then return end
                  local v = meta.typed_metadata["azure_link_id"]
                  if v and #v == 5 then
                    local b1, b2, b3, b4 = v:byte(2), v:byte(3), v:byte(4), v:byte(5)
                    request_handle:headers():replace(
                      "x-azure-link-id",
                      tostring(b1 + b2*256 + b3*65536 + b4*16777216))
                  end
                end
```

{{% /tab %}}
{{< /tabpane >}}

Verify the policy was accepted:

```shell
kubectl get envoypatchpolicy/extract-azure-link-id -o yaml
```

Once traffic flows through a Private Endpoint, the backend will receive an `x-azure-link-id`
header whose value is the decimal `linkIdentifier` (e.g. `4698032`). This value matches the
`linkIdentifier` shown by the Azure control plane for that Private Endpoint connection.

## Preventing header spoofing

Both Lua filters use `headers():replace()`, which overwrites any client-provided value when the
TLV is present. However, when traffic arrives without the expected TLV (for example, a direct
connection that bypasses Private Link), the Lua filter returns early without touching the header,
and a client-supplied value would reach the backend unmodified.

To guard against this, add `earlyRequestHeaders.remove` to the same `ClientTrafficPolicy` that
enables Proxy Protocol. The removal runs before the HTTP filter chain, ensuring the headers are
always stripped from incoming requests before the Lua filter sets them from verified TLV data:

```yaml
headers:
  earlyRequestHeaders:
    remove:
    - x-aws-vpce-id
    - x-azure-link-id
```

## Clean-Up

```shell
kubectl delete envoypatchpolicy/extract-aws-vpce-id
kubectl delete envoypatchpolicy/extract-azure-link-id
```

## See Also

- [Envoy Patch Policy][] — general reference for the `EnvoyPatchPolicy` API
- [Client Traffic Policy][ClientTrafficPolicy] — enabling Proxy Protocol on a Gateway listener
- [Lua Extensions][] — using Lua for custom request/response processing via `EnvoyExtensionPolicy`

[EnvoyPatchPolicy]: ../../../api/extension_types#envoypatchpolicy
[Envoy Patch Policy]: ../extensibility/envoy-patch-policy
[ClientTrafficPolicy]: ../../../api/extension_types#clienttrafficpolicy
[Lua Extensions]: ../extensibility/lua
[aws-pp-spec]: https://docs.aws.amazon.com/elasticloadbalancing/latest/network/edit-target-group-attributes.html#proxy-protocol
[azure-pp-spec]: https://learn.microsoft.com/en-us/azure/private-link/private-link-service-overview#getting-connection-information-using-tcp-proxy-v2
[xds-name-scheme-v2]: ../extensibility/envoy-patch-policy#xds-name-scheme-v2
