Added `BackendTrafficPolicy.spec.http1` (`BackendHTTP1Settings`) for configuring HTTP/1 settings
on upstream backend connections, including `enableTrailers` and `preserveHeaderCase`. When both
`BackendTrafficPolicy.spec.http1` and `ClientTrafficPolicy.spec.http1` are set, `BackendTrafficPolicy`
takes precedence for backend traffic. Setting `http1: {}` (empty) forces explicit HTTP/1 on the
upstream cluster