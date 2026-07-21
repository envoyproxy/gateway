Tracing client sampling now defaults to 0% instead of 100%, so Envoy Gateway no longer honors client-forced tracing unless users explicitly opt in by setting `clientSamplingFraction`.
