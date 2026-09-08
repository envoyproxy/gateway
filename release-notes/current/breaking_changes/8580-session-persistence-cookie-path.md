Cookie-based session persistence now always uses `Path=/` instead of the matched HTTPRoute path, following updated GEP-1619 guidance. This fixes persistence when an upstream proxy rewrites the request path.

Rules on the same host must use distinct `sessionName` values to avoid cookie conflicts, or omit `sessionName` to generate unique names per rule. Existing cookies remain until they expire; clients receive a new cookie with `Path=/` on their next request.
