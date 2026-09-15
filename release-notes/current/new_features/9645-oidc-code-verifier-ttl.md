Added `spec.oidc.codeVerifierTTL` to SecurityPolicy to configure how long the PKCE code verifier cookie
generated during the OAuth2 authorization flow remains valid. Alongside `csrfTokenTTL`, this bounds how
long the cookies of an abandoned authorization flow live in the browser. Defaults to 10 minutes.
