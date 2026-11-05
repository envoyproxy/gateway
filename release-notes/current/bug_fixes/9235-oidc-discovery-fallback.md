Fixed SecurityPolicy OIDC returning HTTP 500 on every protected route when the issuer's
`/.well-known/openid-configuration` could not be fetched during a reconciliation, by retaining
the last successfully discovered configuration per issuer across translations and falling back to
it when discovery fails.
