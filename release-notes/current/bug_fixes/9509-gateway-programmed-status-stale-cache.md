Fixed Gateway status updates getting stuck with `Programmed=False` after Envoy replicas recover by comparing status changes against live API server state instead of a stale informer cache.
