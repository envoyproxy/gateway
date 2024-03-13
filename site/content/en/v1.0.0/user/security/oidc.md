---
title: "OIDC Authentication"
---

This guide provides instructions for configuring [OpenID Connect (OIDC)][oidc] authentication.
OpenID Connect (OIDC) is an authentication standard built on top of OAuth 2.0.
It enables client applications to rely on authentication that is performed by an OpenID Connect Provider (OP)
to verify the identity of a user.

Envoy Gateway introduces a new CRD called [SecurityPolicy][SecurityPolicy] that allows the user to configure OIDC
authentication.
This instantiated resource can be linked to a [Gateway][Gateway] and [HTTPRoute][HTTPRoute] resource.

## Prerequisites

Follow the steps from the [Quickstart](../../quickstart) guide to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

Verify the Gateway status:

```shell
kubectl get gateway/eg -o yaml
```

OIDC can be configured at the Gateway level to authenticate all the HTTPRoutes that are associated with the Gateway with
the same OIDC configuration, or at the HTTPRoute level to authenticate each HTTPRoute with different OIDC configurations. 

This guide demonstrates the configuration of OIDC at the HTTPRoute level.

Let's create an HTTPRoute that represents an application protected by OIDC.

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

Verify the HTTPRoute status:

```shell
kubectl get httproute/myapp -o yaml
```

## Configuration

This guide uses Google as the OIDC provider to demonstrate the configuration of OIDC. However, EG works with any OIDC
providers, including Auth0, Azure AD, Keycloak, Okta, OneLogin, Salesforce, UAA, etc.

### Register an OIDC application

Follow the steps in the [Google OIDC documentation][google-oidc] to register an OIDC application. Please make sure the
redirect URL is set to the one you configured in the SecurityPolicy that you will create in the step below. In this example,
the redirect URL is `http://www.example.com:8080/myapp/oauth2/callback`.

After registering the application, you should have the following information:
* Client ID: The client ID of the OIDC application.
* Client Secret: The client secret of the OIDC application.

### Create a kubernetes secret

Next, create a kubernetes secret with the Client Secret created in the previous step. The secret is an Opaque secret,
and the Client Secret must be stored in the key "client-secret".

Note: please replace the ${CLIENT_SECRET} with the actual Client Secret that you got from the previous step.

```shell
$ kubectl create secret generic my-app-client-secret --from-literal=client-secret=${CLIENT_SECRET}
secret "my-app-client-secret" created
```

### Create a SecurityPolicy

Please notice that the `redirectURL` and `logoutPath` must match the target HTTPRoute. In this example, the target
HTTPRoute is configured to match the host `www.example.com` and the path `/myapp`, so the `redirectURL` must be prefixed 
with `http://www.example.com:8080/myapp`, and `logoutPath` must be prefixed with`/myapp`, otherwise the OIDC authentication 
will fail because the redirect and logout requests will not match the target HTTPRoute and therefore can't be processed 
by the OAuth2 filter on that HTTPRoute.

Note: please replace the ${CLIENT_ID} in the below yaml snippet with the actual Client ID that you got from the OIDC provider.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: oidc-example
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: myapp
  oidc:
    provider:
      issuer: "https://accounts.google.com"
    clientID: "${CLIENT_ID}"
    clientSecret:
      name: "my-app-client-secret"
    redirectURL: "http://www.example.com:8080/myapp/oauth2/callback"
    logoutPath: "/myapp/logout"
EOF
```

Verify the SecurityPolicy configuration:

```shell
kubectl get securitypolicy/oidc-example -o yaml
```

## Testing

Port forward gateway port to localhost:

```shell
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')

kubectl -n envoy-gateway-system port-forward service/${ENVOY_SERVICE} 8080:80
```

Put www.example.com in the /etc/hosts file in your test machine, so we can use this host name to access the gateway from a browser:

```shell
...
127.0.0.1 www.example.com
```

Open a browser and navigate to the `http://www.example.com:8080/myapp` address. You should be redirected to the Google 
login page. After you successfully login, you should see the response from the backend service.

## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) guide to uninstall Envoy Gateway and the example manifest.

Delete the SecurityPolicy, the secret and the HTTPRoute:

```shell
kubectl delete securitypolicy/oidc-example
kubectl delete secret/my-app-client-secret
kubectl delete httproute/myapp
```

## Next Steps

Checkout the [Developer Guide](../../../contributions/develop/) to get involved in the project.

[oidc]: https://openid.net/connect/
[google-oidc]: https://developers.google.com/identity/protocols/oauth2/openid-connect
[SecurityPolicy]: ../../contributions/design/security-policy/
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
