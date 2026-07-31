Fixed unreferenced Secret events triggering a full reconciliation whenever the
HTTPRouteFilter CRD is installed. Every Secret write in the cluster previously
enqueued a reconcile, causing sustained reconcile storms on clusters with
high-frequency Secret writers (secret sync controllers, certificate rotation).
