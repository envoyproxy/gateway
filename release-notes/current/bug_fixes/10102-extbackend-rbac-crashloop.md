Fixed the controller crash-looping when an extension manager declares
`backendResources` for a CRD that is installed but that the controller's
ServiceAccount has no permission to list or watch. The existing guard resolved
through discovery, which every authenticated ServiceAccount is granted, so it
confirmed the Kind was served without confirming it could be read; the watch was
registered anyway, the informer's initial LIST was rejected, and cache sync timed
out for every watched kind. Such a resource is now skipped with a log line, the
same way a missing CRD already was, so reconciliation of unrelated routes
continues.
