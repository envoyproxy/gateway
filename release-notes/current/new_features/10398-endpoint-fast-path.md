Added the experimental `EndpointFastPath` runtime flag, which pushes Kubernetes
endpoint changes to Envoy without waiting for full route and policy translation.
This speeds up endpoint propagation significantly in large deployments: in a test
where full translation took about 7 seconds, endpoint changes reached Envoy in under
100 ms with the flag on, compared with 7 to 15 seconds without it. The flag is disabled by default and
is planned to be enabled by default in the next release. We encourage you to try it,
especially if your Services scale often or roll out frequently, and to report any
issues you run into.
