Reduced inotify watch usage for `BackendTLSPolicy` with `WellKnownCACertificates: System` by sharing a single SDS secret across all policies instead of creating one per policy.
