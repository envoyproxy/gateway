Rate-limited 429s now include `Retry-After` by default.
Set `ClientTrafficPolicy.spec.headers.disableRetryAfterHeader`: true to preserve the previous behavior.
