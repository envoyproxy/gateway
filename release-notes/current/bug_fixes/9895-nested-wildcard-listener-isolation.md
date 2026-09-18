Fixed a route with a concrete hostname attaching to a less specific wildcard listener when a nested
wildcard listener on the same port already owned that hostname, so both listeners served the route.
The listener isolation filter removed a sibling listener's hostname from the route's matched set as a
literal string, which only deleted anything when the set happened to hold that exact string: a concrete
route hostname such as `test.dev.example.com` under a `*.dev.example.com` sibling was never removed, and
the apex `*.example.com` listener kept it. Envoy Gateway now removes any matched hostname that a sibling
listener matches at least as specifically as the listener being computed, ranking an exact hostname above
any wildcard, a longer wildcard suffix above a shorter one, and an empty listener hostname below both.
The failure was invisible in status - every listener stayed `Programmed: True`, the correct certificate
was served on the HTTPS variant, and nothing was logged - so the only reliable detection was comparing
each listener's `rds.route_config_name` against the names actually present in `RoutesConfigDump`.
