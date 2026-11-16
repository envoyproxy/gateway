Fixed the provider runner replacing the whole status of policies, which removed the ancestors written by other controllers. Envoy Gateway now only updates its own entries in `status.ancestors`.
