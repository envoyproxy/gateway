Added `xdsServer.maxReceiveMessageSize` to the EnvoyGateway API and raised the xDS gRPC server's default receive limit from 4MiB to 32MiB. At large resource counts, an Envoy proxy's delta xDS request on stream reconnect can exceed 4MiB, which previously broke the stream with "received message larger than max" errors and left the proxy on its last known-good configuration.

