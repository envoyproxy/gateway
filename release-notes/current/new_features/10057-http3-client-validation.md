HTTP/3 listeners now support client certificate validation. Setting `spec.tls.clientValidation`
together with `spec.http3` on a ClientTrafficPolicy no longer disables HTTP/3; the QUIC listener
requests and validates client certificates the same way the TCP listener does. Requires Envoy
Proxy v1.40 or later.
