The `ClientTrafficPolicy` `http1` fields `enableTrailers`, `preserveHeaderCase`, and `http10`
currently configure both client and backend HTTP/1 protocol settings. In a future release these
fields will be restricted to client connections only. The `ApplyClientTrafficPolicyHTTP1SettingsToClientsOnly`
runtime flag can be used to opt in to the new behavior today.

Migration guidance per field:

- **`enableTrailers`**: use `BackendTrafficPolicy.http1.enableTrailers` for backend settings.
- **`preserveHeaderCase`**: use `BackendTrafficPolicy.http1.preserveHeaderCase` for backend settings.
- **`http10`**: this field currently forces explicit HTTP/1 on backend connections when
  `enableTrailers` and `preserveHeaderCase` are absent. In a future release it will no longer
  affect backend connections. Use `BackendTrafficPolicy.http1: {}` to force explicit HTTP/1 on
  a backend cluster. Note: `accept_http_10` and `default_host_for_http_10` are downstream-only
  fields in Envoy and have never had any effect on upstream clusters.
