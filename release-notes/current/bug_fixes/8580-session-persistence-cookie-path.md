Fixed cookie-based session persistence breaking when the path the client requests differs from the
path Envoy Gateway matches, for example when an edge proxy rewrites `/` to `/foo/bar` before the
request reaches the gateway. The session cookie was scoped to the matched HTTPRoute path (`Path=/foo/bar`),
so the browser never sent it back on the client-visible URL and every request landed on a new backend.
Following the updated GEP-1619 guidance, the session persistence cookie is now always issued with `Path=/`.
Note this is broader than before: the cookie is sent to every route on the same host, not only the one
that set it. Clients holding a cookie scoped to the old path keep using it until it expires, and are
issued a new one at `Path=/` on their next request.
