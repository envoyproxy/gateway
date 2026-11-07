# Simple Extension Server

This example is a simple extension server for e2e testing purposes. For a detailed example of an extension server that leverages custom resources and extension hook context, see the `extension-server` directory.

The extension server modifies virtual hosts. If the vhost name contains the string "fail", the extension server returns an error. Otherwise, it adds another domain to the vhost domain list.

With `--inference-pool`, the server also handles `InferencePool` (`inference.networking.k8s.io/v1`) custom backends in the `Cluster` hook. A cluster built from an InferencePool backendRef is turned into an `ORIGINAL_DST` cluster that sends each request to the `host:port` in the `x-gateway-destination-endpoint` request header, as an endpoint picker would. The `simple-extension-server.envoyproxy.io/destination-header` annotation on a pool names a different header for that pool. There is no endpoint picker, so the client must set the header itself, and this only tests how Envoy Gateway builds and routes custom backend clusters. A cluster hook call that receives more than one InferencePool returns an error. To use it, register `InferencePool` under `extensionManager.backendResources`, add `Cluster` to the post hooks, and pass `--inference-pool` to the container.

[simple-extension-server-inference-pool.yaml](simple-extension-server-inference-pool.yaml) deploys the server with this enabled. The InferencePool CRD must be installed separately.

This server is used in the following tests:
- test/resilience/tests/extensionserver.go
- test/resilience/tests/inference_pool.go (with `--inference-pool`)

