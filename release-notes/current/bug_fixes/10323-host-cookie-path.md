Fixed the OIDC nonce and PKCE code verifier cookies being dropped by the browser when
`cookieNames.oauthNonce` or `cookieNames.codeVerifier` uses the `__Host-` prefix: those two
cookies are now set with `Path=/` for that name instead of the redirect path, so the flow
reaches the callback.
