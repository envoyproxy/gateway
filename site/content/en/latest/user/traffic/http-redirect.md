---
title: "HTTP Redirects"
---

The [HTTPRoute][] resource can issue redirects to clients or rewrite paths sent upstream using filters. Note that
HTTPRoute rules cannot use both filter types at once. Currently, Envoy Gateway only supports __core__
[HTTPRoute filters][] which consist of `RequestRedirect` and `RequestHeaderModifier` at the time of this writing. To
learn more about HTTP routing, refer to the [Gateway API documentation][].

## Prerequisites

Follow the steps from the [Quickstart](../../quickstart) to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTPS.

## Redirects

Redirects return HTTP 3XX responses to a client, instructing it to retrieve a different resource. A
[`RequestRedirect` filter][req_filter] instructs Gateways to emit a redirect response to requests that match the rule.
For example, to issue a permanent redirect (301) from HTTP to HTTPS, configure `requestRedirect.statusCode=301` and
`requestRedirect.scheme="https"`:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-to-https-filter-redirect
spec:
  parentRefs:
    - name: eg
  hostnames:
    - redirect.example
  rules:
    - filters:
      - type: RequestRedirect
        requestRedirect:
          scheme: https
          statusCode: 301
          hostname: www.example.com
          port: 443
EOF
```

__Note:__ `301` (default) and `302` are the only supported statusCodes.

The HTTPRoute status should indicate that it has been accepted and is bound to the example Gateway.

```shell
kubectl get httproute/http-to-https-filter-redirect -o yaml
```

Get the Gateway's address:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Querying `redirect.example/get` should result in a `301` response from the example Gateway and redirecting to the
configured redirect hostname.

```console
$ curl -L -vvv --header "Host: redirect.example" "http://${GATEWAY_HOST}/get"
...
< HTTP/1.1 301 Moved Permanently
< location: https://www.example.com/get
...
```

If you followed the steps in the [Secure Gateways](../security/secure-gateways) guide, you should be able to curl the redirect
location.

## HTTP --> HTTPS

Listeners expose the TLS setting on a per domain or subdomain basis. TLS settings of a listener are applied to all domains that satisfy the hostname criteria.

Create a root certificate and private key to sign certificates:

```shell
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/CN=example.com' -keyout CA.key -out CA.crt
openssl req -out example.com.csr -newkey rsa:2048 -nodes -keyout tls.key -subj "/CN=example.com"
```

Generate a self-signed wildcard certificate for `example.com` with `*.example.com` extension

```shell
cat <<EOF | openssl x509 -req -days 365 -CA CA.crt -CAkey CA.key -set_serial 0 \
-subj "/CN=example.com" \
-in example.com.csr -out tls.crt -extensions v3_req  -extfile -
[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1   = example.com
DNS.2   = *.example.com
EOF
```

Create the kubernetes tls secret

```shell
kubectl create secret tls example-com --key=tls.key --cert=tls.crt
```

Define a https listener on the existing gateway

```shell
cat <<EOF | kubectl apply -n default -f -
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: eg
spec:
  gatewayClassName: eg
  listeners:
  - name: http
    port: 80
    protocol: HTTP
    # hostname: "*.example.com"
  - name: https
    port: 443
    protocol: HTTPS
    # hostname: "*.example.com"
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        name: example-com
EOF
```

Check for any TLS certificate issues on the gateway.

```bash
kubectl -n default describe gateway eg
```

Create two HTTPRoutes and attach them to the HTTP and HTTPS listeners using the [sectionName][] field.


```shell
cat <<EOF | kubectl apply -n default -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: tls-redirect
spec:
  parentRefs:
    - name: eg
      sectionName: http
  hostnames:
    # - "*.example.com" # catch all hostnames
    - "www.example.com"
  rules:
    - filters:
        - type: RequestRedirect
          requestRedirect:
            scheme: https
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: backend
spec:
  parentRefs:
    - name: eg
      sectionName: https
  hostnames:
    - "www.example.com"
  rules:
    - backendRefs:
        - group: ""
          kind: Service
          name: backend
          port: 3000
          weight: 1
      matches:
        - path:
            type: PathPrefix
            value: /
EOF
```

Curl the example app through http listener:

```bash
curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST/get
```

Curl the example app through https listener:

```bash
curl -v -H 'Host:www.example.com' --resolve "www.example.com:443:$GATEWAY_HOST" \
--cacert CA.crt https://www.example.com:443/get
```


## Path Redirects

Path redirects use an HTTP Path Modifier to replace either entire paths or path prefixes. For example, the HTTPRoute
below will issue a 302 redirect to all `path.redirect.example` requests whose path begins with `/get` to `/status/200`.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-filter-path-redirect
spec:
  parentRefs:
    - name: eg
  hostnames:
    - path.redirect.example
  rules:
    - matches:
      - path:
          type: PathPrefix
          value: /get
      filters:
      - type: RequestRedirect
        requestRedirect:
          path:
            type: ReplaceFullPath
            replaceFullPath: /status/200
          statusCode: 302
      backendRefs:
      - name: backend
        port: 3000
EOF
```

The HTTPRoute status should indicate that it has been accepted and is bound to the example Gateway.

```shell
kubectl get httproute/http-filter-path-redirect -o yaml
```

Querying `path.redirect.example` should result in a `302` response from the example Gateway and a redirect location
containing the configured redirect path.

Query the `path.redirect.example` host:

```shell
curl -vvv --header "Host: path.redirect.example" "http://${GATEWAY_HOST}/get"
```

You should receive a `302` with a redirect location of `http://path.redirect.example/status/200`.

[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[HTTPRoute filters]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.HTTPRouteFilter
[Gateway API documentation]: https://gateway-api.sigs.k8s.io/
[req_filter]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.HTTPRequestRedirectFilter
[sectionName]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.CommonRouteSpec
