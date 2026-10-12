ClientTrafficPolicy `http1.http10` settings no longer configure backend HTTP/1.0 options.
The `accept_http_10` and `default_host_for_http_10` fields on `Http1ProtocolOptions` are only
used by Envoy when configured in client (downstream) context. When set, these fields will only
force explicit HTTP/1 for backend connections, when other backend-relevant HTTP/1 fields like 
`enableTrailers` and `preserveHeaderCase` are absent. HTTP/1.0 downstream traffic continues to be 
proxied correctly without these backend settings. 

In a future release, `http1.http10` will no longer force explicit HTTP/1 on backend connections.
To force HTTP/1 on a backend cluster explicitly, use `BackendTrafficPolicy.http1: {}`.
