Fixed OIDC flow-state cookies accumulating in the browser and overflowing the request header size limit.
Envoy mints a nonce (CSRF) and a PKCE code verifier cookie for every authorization flow it starts, but
only deletes the pair belonging to the flow that completes the callback, so flows that are abandoned -
parallel requests from a logged out browser, a user navigating away from the provider's login page -
leave their cookies behind until they expire. Envoy Gateway now scopes both cookies to the OIDC redirect
path, the only path where Envoy needs their value, so any orphans are no longer sent on every request.
Note this bounds the damage rather than eliminating it: orphans are still sent to the callback endpoint
itself until they expire, and logout can no longer purge them early because the browser no longer sends
them to the signout path, so also consider lowering `csrfTokenTTL`.
The PKCE code verifier cookie is also now named `CodeVerifier-<suffix>`, carrying the same per-policy
suffix as the other OAuth2 cookies instead of Envoy's shared default, so SecurityPolicies on the same
cookie domain no longer delete each other's in-flight flow cookies on logout.
On upgrade, a browser already holding flow cookies keeps them at the old `path=/`, since the new
deletion headers are scoped to the redirect path. They expire on the lifetime they were originally
issued with - 10 minutes by default, and unaffected by any `csrfTokenTTL` you configure during the
upgrade.
The old code verifier cookie name is also no longer read, so a login that was in progress across the
rollout may need to be retried once.
