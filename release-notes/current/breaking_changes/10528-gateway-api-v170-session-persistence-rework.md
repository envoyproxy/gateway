Gateway API v1.7.0 reworked the experimental `sessionPersistence` API on
`HTTPRoute` and `GRPCRoute`: the `sessionName` field was removed in favor of
`cookie.name` (cookie-based persistence) and `header.name` (header-based
persistence, now required), and `cookieConfig` was renamed to `cookie`. Update
any manifests using `sessionPersistence` on the experimental channel before
upgrading to the v1.7.0 Gateway API CRDs: replace `sessionName` with
`cookie: {name: <name>}` or `header: {name: <name>}`, and `cookieConfig` with
`cookie`. Cookie-based rules missing `cookie.name` fall back to generated
cookie names and session-lifetime cookies, while header-based rules missing
`header.name` are rejected with an `UnsupportedValue` route condition until
the field is set.
