Added OIDC token cookie chunking: Envoy now splits an access, ID or refresh token too large for one browser cookie
across several cookies, instead of the browser dropping it and the login looping back to the OIDC provider. Tokens
that fit in one cookie are written as before. Envoy Proxy older than v1.40 can't read the split cookies, so while
proxies of an earlier release serve the same hostnames, for example during the upgrade, disable the new
`OIDCTokenCookieChunking` runtime flag, which is enabled by default, through `runtimeFlags.disabled`.
