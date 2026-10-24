Added `spec.oidc.tokenCookieChunking` to SecurityPolicy so that Envoy splits an access, ID or refresh token
too large for one browser cookie across several cookies, instead of the browser dropping it and the login
looping back to the OIDC provider. Optional `maxCookieSize` and `maxChunks` bound the split. Chunking is off
unless set and requires Envoy v1.40 or later.
