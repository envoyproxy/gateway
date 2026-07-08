Added a new `host` section under `ClientTrafficPolicy`'s `headers` with a `stripTrailingHostDot` field to normalize the Host/Authority header (trailing dot removal) without an EnvoyPatchPolicy.
