# Envoy SDS Server Example

This is a simple gRPC server that implements the Envoy Secret Discovery Service (SDS) protocol.

## Overview

The SDS server provides TLS certificates and validation contexts to Envoy proxies dynamically. This example implementation:

- Generates self-signed certificates on startup
- Serves secrets via the SDS gRPC protocol
- Supports both dedicated SDS and ADS (Aggregated Discovery Service)
- Provides two types of secrets:
  - `server_cert`: TLS certificate with private key
  - `validation_context`: Certificate validation context (CA)

## Building

```bash
# From the sds-server directory
go mod download
go build -o sds-server .
```

Or use the provided Makefile:

```bash
make build
```

## Running

Start the SDS server:

```bash
./sds-server
```

Options:
- `-port <port>`: gRPC port for SDS server (default: 18001)
- `-node <nodeID>`: Node ID for envoy client (default: "sds-test-node")
- `-cn <commonName>`: Common Name for the generated certificate (default: "sds-test.example.com")
- `-dns <dnsNames>`: Comma-separated list of DNS names for the certificate (default: "sds-test.example.com,*.example.com,localhost")

Example:

```bash
./sds-server -port 18001 -node my-envoy-node -cn myapp.example.com -dns "myapp.example.com,*.myapp.example.com,localhost"
```

Or with the Makefile:

```bash
make run PORT=18001 NODE_ID=my-envoy-node COMMON_NAME=myapp.example.com DNS_NAMES="myapp.example.com,*.myapp.example.com,localhost"
```

## Testing with Envoy

Here's an example Envoy configuration that uses this SDS server:

```yaml
node:
  id: sds-test-node
  cluster: test-cluster

dynamic_resources:
  cds_config:
    resource_api_version: V3
    api_config_source:
      api_type: GRPC
      transport_api_version: V3
      grpc_services:
      - envoy_grpc:
          cluster_name: sds_cluster

static_resources:
  clusters:
  - name: sds_cluster
    type: STATIC
    connect_timeout: 1s
    http2_protocol_options: {}
    load_assignment:
      cluster_name: sds_cluster
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 18001
  
  listeners:
  - name: listener_0
    address:
      socket_address:
        address: 0.0.0.0
        port_value: 10000
    filter_chains:
    - transport_socket:
        name: envoy.transport_sockets.tls
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.DownstreamTlsContext
          common_tls_context:
            tls_certificate_sds_secret_configs:
            - name: server_cert
              sds_config:
                resource_api_version: V3
                api_config_source:
                  api_type: GRPC
                  transport_api_version: V3
                  grpc_services:
                  - envoy_grpc:
                      cluster_name: sds_cluster
      filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          stat_prefix: ingress_http
          route_config:
            name: local_route
            virtual_hosts:
            - name: backend
              domains: ["*"]
              routes:
              - match: { prefix: "/" }
                direct_response:
                  status: 200
                  body:
                    inline_string: "Hello from Envoy with SDS!\n"
          http_filters:
          - name: envoy.filters.http.router
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router

admin:
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 9901
```

Save this as `envoy-config.yaml` and run:

```bash
envoy -c envoy-config.yaml
```

## Docker

Build the Docker image:

```bash
docker build -t sds-server:latest .
```

Run the container:

```bash
docker run -p 18001:18001 sds-server:latest
```

## Architecture

The server implements the following flow:

1. **Initialization**: Generates a self-signed certificate and private key
2. **Snapshot Creation**: Creates an xDS snapshot containing two secrets:
   - TLS certificate secret (for server authentication)
   - Validation context secret (for client certificate validation)
3. **gRPC Service**: Registers the SecretDiscoveryService and AggregatedDiscoveryService
4. **Streaming**: Handles streaming requests from Envoy proxies
5. **Callbacks**: Logs all discovery requests and responses for debugging

## Features

- ✅ SDS v3 API support
- ✅ Self-signed certificate generation with configurable Common Name and DNS names
- ✅ Snapshot cache for secrets
- ✅ gRPC keepalive configuration
- ✅ Request/response logging
- ✅ Support for both SDS and ADS protocols
- ✅ Node ID-based secret delivery

## Notes

- The server generates a new certificate on each startup
- The certificate is valid for 1 year
- Certificate Common Name and DNS names (SANs) are configurable via command-line flags
- Default DNS names include: `sds-test.example.com`, `*.example.com`, and `localhost`
- The certificate always includes IP address `127.0.0.1`
- In production, you should load certificates from secure storage (e.g., Kubernetes secrets, HashiCorp Vault)

## References

- [Envoy SDS Documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/security/secret)
- [go-control-plane](https://github.com/envoyproxy/go-control-plane)
- [Envoy API Reference](https://www.envoyproxy.io/docs/envoy/latest/api-v3/api)
