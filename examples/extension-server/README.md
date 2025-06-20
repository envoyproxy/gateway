# Extension Server for InferencePool Custom Backend Support

This extension server implements custom backend support for InferencePool resources in Envoy Gateway. When an HTTPRoute uses an InferencePool as a backend reference, this extension server will:

1. Detect the InferencePool custom backend
2. Create two clusters: an `original_destination_cluster` and an `ext_proc_cluster`
3. Configure the original destination cluster to use HTTP header routing with `x-gateway-destination-endpoint` header
4. Modify the route to use the original destination cluster with a 120-second timeout

## Features

- **InferencePool Detection**: Automatically detects InferencePool custom backends using the API version `sigs.k8s.io/gateway-api-inference-extension/v1alpha2`
- **Dual Cluster Creation**: Creates both original destination and external processing clusters
- **Dynamic Cluster Naming**: Uses naming format `endpointpicker_{name}_{namespace}_original_dst` and `endpointpicker_{name}_{namespace}_ext_proc`
- **Original Destination Load Balancing**: Uses ORIGINAL_DST cluster type with HTTP header-based routing
- **External Processing Support**: Creates STRICT_DNS cluster for endpoint picker service
- **Route Modification**: Updates routes to use the original destination cluster with extended timeout

## Implementation Details

### PostRouteModify Method

The `PostRouteModify` method is called by Envoy Gateway after generating a Route xDS configuration. It:

1. **Parses Extension Resources**: Checks for InferencePool resources in the route context using unstructured JSON unmarshaling
2. **Creates Two Custom Clusters**:
   - **Original Destination Cluster**: For routing traffic based on HTTP headers
   - **External Processing Cluster**: For the endpoint picker service

3. **Original Destination Cluster Configuration**:

   ```go
   &clusterv3.Cluster{
       Name: "endpointpicker_{name}_{namespace}_original_dst",
       ClusterDiscoveryType: &clusterv3.Cluster_Type{
           Type: clusterv3.Cluster_ORIGINAL_DST,
       },
       LbPolicy: clusterv3.Cluster_CLUSTER_PROVIDED,
       ConnectTimeout: durationpb.New(6 * time.Second),
       DnsLookupFamily: clusterv3.Cluster_V4_ONLY,
       LbConfig: &clusterv3.Cluster_OriginalDstLbConfig_{
           OriginalDstLbConfig: &clusterv3.Cluster_OriginalDstLbConfig{
               UseHttpHeader: true,
               HttpHeaderName: "x-gateway-destination-endpoint",
           },
       },
   }
   ```

4. **External Processing Cluster Configuration**:

   ```go
   &clusterv3.Cluster{
       Name: "endpointpicker_{name}_{namespace}_ext_proc",
       ClusterDiscoveryType: &clusterv3.Cluster_Type{
           Type: clusterv3.Cluster_STRICT_DNS,
       },
       LbPolicy: clusterv3.Cluster_LEAST_REQUEST,
       ConnectTimeout: durationpb.New(10 * time.Second),
       // LoadAssignment points to the endpoint picker service
   }
   ```

5. **Modifies Route**: Updates the route to use the original destination cluster and sets timeout to 120 seconds

### Cluster Configuration

The extension server creates two clusters for each InferencePool:

#### Original Destination Cluster

```yaml
- type: "type.googleapis.com/envoy.config.cluster.v3.Cluster"
  name: "endpointpicker_{name}_{namespace}_original_dst"
  operation:
    op: add
    path: ""
    value:
      name: endpointpicker_{name}_{namespace}_original_dst
      type: ORIGINAL_DST
      original_dst_lb_config:
        use_http_header: true
        http_header_name: "x-gateway-destination-endpoint"
      connect_timeout: 6s
      lb_policy: CLUSTER_PROVIDED
      dns_lookup_family: V4_ONLY
```

#### External Processing Cluster

```yaml
- type: "type.googleapis.com/envoy.config.cluster.v3.Cluster"
  name: "endpointpicker_{name}_{namespace}_ext_proc"
  operation:
    op: add
    path: ""
    value:
      name: endpointpicker_{name}_{namespace}_ext_proc
      type: STRICT_DNS
      lb_policy: LEAST_REQUEST
      connect_timeout: 10s
      load_assignment:
        cluster_name: endpointpicker_{name}_{namespace}_ext_proc
        endpoints:
        - lb_endpoints:
          - endpoint:
              address:
                socket_address:
                  address: "{extension_service_name}.{namespace}.svc"
                  port_value: {extension_port}
```

### Route Configuration

The route is modified to use the original destination cluster:

```yaml
route:
  cluster: endpointpicker_{name}_{namespace}_original_dst
  timeout: 120s
```

## Usage

1. **Build the Extension Server**:
   ```bash
   go build ./cmd/extension-server
   ```

2. **Deploy the Extension Server**: Deploy the server in your Kubernetes cluster

3. **Configure Envoy Gateway**: Set up the EnvoyProxy resource to use this extension server for PostRouteModify hooks

4. **Create HTTPRoute with InferencePool Backend**: Use InferencePool as a backend reference in your HTTPRoute

## Key Implementation Details

### InferencePool Detection

The server detects InferencePool resources by:

- Checking the `kind` field equals "InferencePool"
- Verifying the `apiVersion` is "sigs.k8s.io/gateway-api-inference-extension/v1alpha2"
- Unmarshaling the unstructured JSON directly to the InferencePool type

### Cluster Naming Convention

- **Original Destination Cluster**: `endpointpicker_{name}_{namespace}_original_dst`
- **External Processing Cluster**: `endpointpicker_{name}_{namespace}_ext_proc`

Where `{name}` and `{namespace}` come from the InferencePool resource metadata.

### HTTP Header Configuration

The original destination cluster uses the `x-gateway-destination-endpoint` header to determine the target endpoint, which differs from the previous `target-pod` header name.

## Testing

Run the tests to verify the implementation:

```bash
go test ./internal/extensionserver -v
```

The tests cover:

- InferencePool detection and processing
- Dual cluster creation (original destination + external processing)
- Route modification with correct cluster reference
- Error handling for invalid resources
- Proper cluster naming conventions

## Dependencies

- `sigs.k8s.io/gateway-api-inference-extension/api/v1alpha2` - InferencePool API types
- Envoy Gateway extension API
- Envoy xDS API (cluster and route v3)
- `k8s.io/apimachinery/pkg/apis/meta/v1/unstructured` - For dynamic resource parsing
