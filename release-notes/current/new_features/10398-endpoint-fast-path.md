Added the disabled-by-default `EndpointFastPath` runtime flag to publish eligible
Kubernetes EndpointSlice changes before full route and policy translation finishes.
Existing Kubernetes watches continue to trigger full reconciliation for relevant
endpoint changes, including updates that require route or cluster changes.
The flag is planned to be enabled by default in the next release. If your clusters
see frequent endpoint changes, please try it on this release and report any issues.
