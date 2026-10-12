ClientTrafficPolicy `http1.http10` settings no longer configure backend HTTP/1.0 options.
The `accept_http_10` and `default_host_for_http_10` fields on `Http1ProtocolOptions` are only
used by Envoy when configured in client (downstream) context. The upstream HTTP/1 codec
unconditionally writes `HTTP/1.1` as the request line protocol — hardcoded as `REQUEST_POSTFIX`
in `source/common/http/http1/codec_impl.cc` — so these fields have no effect on upstream clusters.
HTTP/1.0 downstream traffic continues to be proxied correctly without these backend settings.
