Added the disabled-by-default `EndpointFastPath` runtime flag to publish eligible
Kubernetes EndpointSlice changes before full route and policy translation finishes.
Full reconciliation remains enabled for every endpoint change and handles updates
that require route or cluster changes.
