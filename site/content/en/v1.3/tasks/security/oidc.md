---
title: "OIDC Authentication"
---

This task provides instructions for configuring [OpenID Connect (OIDC)][oidc] authentication.
OpenID Connect (OIDC) is an authentication standard built on top of OAuth 2.0.
It enables EG to rely on authentication that is performed by an OpenID Connect Provider (OP)
to verify the identity of a user.

Envoy Gateway introduces a new CRD called [SecurityPolicy][SecurityPolicy] that allows the user to configure OIDC
authentication.
This instantiated resource can be linked to a [Gateway][Gateway] and [HTTPRoute][HTTPRoute] resource.

## Prerequisites

{{< boilerplate prerequisites >}}

EG OIDC authentication requires the redirect URL to be HTTPS. Follow the [Secure Gateways](../secure-gateways) guide
to generate the TLS certificates and update the Gateway configuration to add an HTTPS listener.

Verify the Gateway status:

```shell
kubectl get gateway/eg -o yaml
```

Let's create an HTTPRoute that represents an application protected by OIDC.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: myapp
spec:
  parentRefs:
  - name: eg
  hostnames: ["www.example.com"]
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /myapp
    backendRefs:
    - name: backend
      port: 3000
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: myapp
spec:
  parentRefs:
  - name: eg
  hostnames: ["www.example.com"]
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /myapp
    backendRefs:
    - name: backend
      port: 3000
```

{{% /tab %}}
{{< /tabpane >}}

Verify the HTTPRoute status:

```shell
kubectl get httproute/myapp -o yaml
```

## OIDC Authentication for a HTTPRoute

OIDC can be configured at the Gateway level to authenticate all the HTTPRoutes that are associated with the Gateway with
the same OIDC configuration, or at the HTTPRoute level to authenticate each HTTPRoute with different OIDC configurations.

This section demonstrates how to configure OIDC authentication for a specific HTTPRoute.

### Register an OIDC application

This task uses Google as the OIDC provider to demonstrate the configuration of OIDC. However, EG works with any OIDC
providers, including Auth0, Azure AD, Keycloak, Okta, OneLogin, Salesforce, UAA, etc.

Follow the steps in the [Google OIDC documentation][google-oidc] to register an OIDC application. Please make sure the
redirect URL is set to the one you configured in the SecurityPolicy that you will create in the step below. In this example,
the redirect URL is `https://www.example.com:8443/myapp/oauth2/callback`.

After registering the application, you should have the following information:
* Client ID: The client ID of the OIDC application.
* Client Secret: The client secret of the OIDC application.

### Create a kubernetes secret

Next, create a kubernetes secret with the Client Secret created in the previous step. The secret is an Opaque secret,
and the Client Secret must be stored in the key "client-secret".

Note: please replace the ${CLIENT_SECRET} with the actual Client Secret that you got from the previous step.

```shell
kubectl create secret generic my-app-client-secret --from-literal=client-secret=${CLIENT_SECRET}
```

### Create a SecurityPolicy

**Please notice that the `redirectURL` and `logoutPath` must match the target HTTPRoute.** In this example, the target
HTTPRoute is configured to match the host `www.example.com` and the path `/myapp`, so the `redirectURL` must be prefixed
with `https://www.example.com:8443/myapp`, and `logoutPath` must be prefixed with`/myapp`, otherwise the OIDC authentication
will fail because the redirect and logout requests will not match the target HTTPRoute and therefore can't be processed
by the OAuth2 filter on that HTTPRoute.

Note: please replace the ${CLIENT_ID} in the below yaml snippet with the actual Client ID that you got from the OIDC provider.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: oidc-example
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: myapp
  oidc:
    provider:
      issuer: "https://accounts.google.com"
    clientID: "${CLIENT_ID}"
    clientSecret:
      name: "my-app-client-secret"
    redirectURL: "https://www.example.com:8443/myapp/oauth2/callback"
    logoutPath: "/myapp/logout"
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: oidc-example
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: myapp
  oidc:
    provider:
      issuer: "https://accounts.google.com"
    clientID: "${CLIENT_ID}"
    clientSecret:
      name: "my-app-client-secret"
    redirectURL: "https://www.example.com:8443/myapp/oauth2/callback"
    logoutPath: "/myapp/logout"
```

{{% /tab %}}
{{< /tabpane >}}

Verify the SecurityPolicy configuration:

```shell
kubectl get securitypolicy/oidc-example -o yaml
```

### Testing

Port forward gateway port to localhost:

```shell
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')

kubectl -n envoy-gateway-system port-forward service/${ENVOY_SERVICE} 8443:443
```

Put www.example.com in the /etc/hosts file in your test machine, so we can use this host name to access the gateway from a browser:

```shell
...
127.0.0.1 www.example.com
```

Open a browser and navigate to the `https://www.example.com:8443/myapp` address. You should be redirected to the Google
login page. After you successfully login, you should see the response from the backend service.

Clean the cookies in the browser and try to access `https://www.example.com:8443/foo` address. You should be able to see
this page since the path `/foo` is not protected by the OIDC policy.

## OIDC Authentication for a Gateway

OIDC can be configured at the Gateway level to authenticate all the HTTPRoutes that are associated with the Gateway with
the same OIDC configuration, or at the HTTPRoute level to authenticate each HTTPRoute with different OIDC configurations.

This section demonstrates how to configure OIDC authentication for a Gateway.

### Register an OIDC application

If you haven't registered an OIDC application, follow the steps in the previous section to register an OIDC application.

### Create a kubernetes secret

If you haven't created a kubernetes secret, follow the steps in the previous section to create a kubernetes secret.

### Create an HTTPRoute with a different subdomain

Let's create another HTTPRoute in the same Gateway, but with a different subdomain.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: foo
spec:
  parentRefs:
  - name: eg
  hostnames: ["foo.example.com"]
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - name: backend
      port: 3000
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: foo
spec:
  parentRefs:
  - name: eg
  hostnames: ["foo.example.com"]
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /foo
    backendRefs:
    - name: backend
      port: 3000
```

{{% /tab %}}
{{< /tabpane >}}

Verify the HTTPRoute status:

```shell
kubectl get httproute/foo -o yaml
```

### Create a SecurityPolicy

Create or update the SecurityPolicy to target the Gateway instead of the HTTPRoute. **Please notice that the `redirectURL`
and `logoutPath` must match one of the HTTPRoutes associated with the Gateway.** In this example, the target Gateway has
three HTTPRoutes associated with it, one with the host `www.example.com` and the path `/myapp`, one with the host
`www.example.com` and the path `/`, and one with the host `foo.example.com` and the path `/`. Any of these HTTPRoutes
can be used to match the `redirectURL` and `logoutPath`.

By default, the access token and ID token cookies are set to the host of the request, excluding subdomains. To allow the
token cookies to be shared across subdomains and prevent users from having to log in again when switching between subdomains,
the `cookieDomain` field needs to be set to the root domain. In this example, the root domain is `example.com`.

Note: if a `cookieDomain` is added to an existing SecurityPolicy, the cookies in the browser must be cleared before sending a new request to the Gateway, otherwise the cookies with the old subdomain will take precedence and be sent to the Gateway, causing the OIDC authentication to fail.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: oidc-example
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  oidc:
    provider:
      issuer: "https://accounts.google.com"
    clientID: "${CLIENT_ID}"
    clientSecret:
      name: "my-app-client-secret"
    redirectURL: "https://www.example.com:8443/myapp/oauth2/callback"
    logoutPath: "/myapp/logout"
    cookieDomain: "example.com"
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: oidc-example
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  oidc:
    provider:
      issuer: "https://accounts.google.com"
    clientID: "${CLIENT_ID}"
    clientSecret:
      name: "my-app-client-secret"
    redirectURL: "https://www.example.com:8443/myapp/oauth2/callback"
    logoutPath: "/myapp/logout"
    cookieDomain: "example.com"
```

{{% /tab %}}
{{< /tabpane >}}

Verify the SecurityPolicy configuration:

```shell
kubectl get securitypolicy/oidc-example -o yaml
```

### Update the Listener TLS certificate to support multiple subdomains

Create a multi-domain wildcard certificate for `*.example.com`.

```shell
openssl req -out wildcard.csr -newkey rsa:2048 -nodes -keyout wildcard.key -subj "/CN=*.example.com/O=example organization"
openssl x509 -req -days 365 -CA example.com.crt -CAkey example.com.key -set_serial 0 -in wildcard.csr -out wildcard.crt
```

Replace the TLS certificate of the Gateway with the wildcard certificate.

```shell
kubectl delete secret example-cert
kubectl create secret tls example-cert --key=wildcard.key --cert=wildcard.crt
```

### Testing

If you haven't done so, follow the steps in the previous section to port forward gateway port to localhost and put
www.example.com in the /etc/hosts file in your test machine.

Also, put foo.example.com in the /etc/hosts file in your test machine.

```shell
...
127.0.0.1 foo.example.com
```

Open a browser and navigate to the `https://www.example.com:8443/myapp` address. You should be redirected to the Google
login page. After you successfully login, you should see the response from the backend service.

You can also try to access `https://foo.example.com:8443` and `https://www.example.com:8443/bar` addresses. You should
be able to see the response from the backend service since these HTTPRoutes are also protected by the same OIDC config,
and the cookies are shared across subdomains.

## Connect to an OIDC Provider with Self-Signed Certificate

In some scenarios, the OIDC provider may use a self-signed certificate. To connect to an OIDC provider with a self-signed certificate, you need to configure it using the [Backend] resource within the [SecurityPolicy]. Additionally, use the [BackendTLSPolicy] to specify the CA certificate required to authenticate the OIDC provider.

The following example demonstrates how to configure the OIDC provider with a self-signed certificate.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: oidc-example
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: myapp
  oidc:
    provider:
      backendRefs:
      - group: gateway.envoyproxy.io
        kind: Backend
        name: backend-keycloak
        port: 443
      backendSettings:
        retry:
          numRetries: 3
          perRetry:
            backOff:
              baseInterval: 1s
              maxInterval: 5s
          retryOn:
            triggers: ["5xx", "gateway-error", "reset"]
      issuer: "https://my.keycloak.com/realms/master"
      authorizationEndpoint: "https://my.keycloak.com/realms/master/protocol/openid-connect/auth"
      tokenEndpoint: "https://my.keycloak.com/realms/master/protocol/openid-connect/token"
    clientID: "${CLIENT_ID}"
    clientSecret:
      name: "my-app-client-secret"
    redirectURL: "http://www.example.com/myapp/oauth2/callback"
    logoutPath: "/myapp/logout"
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: backend-keycloak
spec:
  endpoints:
  - fqdn:
      hostname: 'my.keycloak.com'
      port: 443
---
apiVersion: gateway.networking.k8s.io/v1alpha3
kind: BackendTLSPolicy
metadata:
  name: policy-btls
spec:
  targetRefs:
  - group: gateway.envoyproxy.io
    kind: Backend
    name: backend-keycloak
  validation:
    caCertificateRefs:
    - name: backend-tls-certificate
      group: ""
      kind: ConfigMap
    hostname: my.keycloak.com
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: oidc-example
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: myapp
  oidc:
    provider:
      backendRefs:
      - group: gateway.envoyproxy.io
        kind: Backend
        name: backend-keycloak
        port: 443
      backendSettings:
        retry:
          numRetries: 3
          perRetry:
            backOff:
              baseInterval: 1s
              maxInterval: 5s
          retryOn:
            triggers: ["5xx", "gateway-error", "reset"]
      issuer: "https://my.keycloak.com/realms/master"
      authorizationEndpoint: "https://my.keycloak.com/realms/master/protocol/openid-connect/auth"
      tokenEndpoint: "https://my.keycloak.com/realms/master/protocol/openid-connect/token"
    clientID: "${CLIENT_ID}"
    clientSecret:
      name: "my-app-client-secret"
    redirectURL: "http://www.example.com/myapp/oauth2/callback"
    logoutPath: "/myapp/logout"
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: backend-keycloak
spec:
  endpoints:
  - fqdn:
      hostname: 'my.keycloak.com'
      port: 443
---
apiVersion: gateway.networking.k8s.io/v1alpha3
kind: BackendTLSPolicy
metadata:
  name: policy-btls
spec:
  targetRefs:
  - group: gateway.envoyproxy.io
    kind: Backend
    name: backend-keycloak
  validation:
    caCertificateRefs:
    - name: backend-tls-certificate
      group: ""
      kind: ConfigMap
    hostname: my.keycloak.com
```

{{% /tab %}}
{{< /tabpane >}}

As shown in the example above, the [SecurityPolicy] resource is configured with an OIDC provider in its OIDC settings. The `backendRefs` field references the [Backend] resource that defines the OIDC provider. The [BackendTLSPolicy] resource specifies the CA certificate required to authenticate the OIDC provider.

Additional connection settings for the OIDC provider can be configured in the [backendSettings]. Currently, only the retry policy is supported.

For more information about [Backend] and [BackendTLSPolicy], refer to the [Backend Routing][backend-routing] and [Backend TLS: Gateway to Backend][backend-tls] tasks.

## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the SecurityPolicy, the secret and the HTTPRoute:

```shell
kubectl delete securitypolicy/oidc-example
kubectl delete secret/my-app-client-secret
kubectl delete httproute/myapp
kubectl delete httproute/foo
```

## Next Steps

Checkout the [Developer Guide](../../../../contributions/develop) to get involved in the project.

[oidc]: https://openid.net/connect/
[google-oidc]: https://developers.google.com/identity/protocols/oauth2/openid-connect
[SecurityPolicy]: ../../../api/extension_types#securitypolicy
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
[Backend]: ../../../api/extension_types#backend
[BackendTLSPolicy]: https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/
[backend-routing]: ../traffic/backend
[backend-tls]: ../backend-tls
[BackendSettings]: ../../../api/extension_types/#clustersettings
