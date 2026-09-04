The cookie-based session persistence cookie is now always issued with `Path=/`, following the updated
GEP-1619 guidance, instead of being scoped to the matched HTTPRoute path. This fixes session persistence
breaking when the path the client requests differs from the path Envoy Gateway matches, for example when
an edge proxy rewrites `/` to `/foo/bar` before the request reaches the gateway: the cookie was scoped to
`/foo/bar`, so the browser never sent it back on the client-visible URL.
Because the cookie is now sent to every route on the same host, HTTPRoute rules on the same host that
enable cookie-based session persistence must use distinct `sessionName` values. Rules that share a
`sessionName` previously got separate cookies because their paths differed; they now overwrite each other's
cookie and session persistence stops working for both. Clients holding a cookie scoped to the old path keep
sending it until it expires, and are issued a new one at `Path=/` on their next request.
