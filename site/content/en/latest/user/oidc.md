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

Follow the steps from the [Quickstart](../quickstart) guide to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

OIDC authentication requires the redirect URL to be HTTPS. Follow the [Secure Gateways](../secure-gateways) guide
 to generate the TLS certificates and update the Gateway configuration to add an HTTPS listener.

Verify the Gateway status:

```shell
kubectl get gateway/teg -o yaml
```

## Configuration

This guide uses Google as the OIDC provider to demonstrate the configuration of OIDC. However, EG works with any OIDC
providers, including Auth0, Azure AD, Keycloak, Okta, OneLogin, Salesforce, UAA, etc.

### Register an OIDC application

Follow the steps in the [Google OIDC documentation][google-oidc] to register an OIDC application. Please make sure the
redirect URL is set to the one you configured in the SecurityPolicy that you will create in the step below. If you don't
specify a redirect URL in the SecurityPolicy, the default redirect URL is `https://${GATEWAY_HOST}/oauth2/callback`.
Please notice that the `redirectURL` and `logoutPath` must be caught by the target HTTPRoute. For example, if the 
HTTPRoute is configured to match the host `www.example.com` and the path `/foo`, the `redirectURL` must
be prefixed with `https://www.example.com/foo`, and `logoutPath` must be prefixed with`/foo`, for example,
`https://www.example.com/foo/oauth2/callback` and `/foo/logout`, otherwise the OIDC authentication will fail.

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

Note: please replace the ${CLIENT_ID} with the actual Client ID that you got from the previous step.

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
    name: backend
  oidc:
    provider:
      issuer: "https://accounts.google.com"
    clientID: "${CLIENT_ID}.apps.googleusercontent.com"
    clientSecret:
      name: "my-app-client-secret"
    redirectURI: "https://www.example.com/oauth2/callback"
    logoutPath: "/logout"
EOF
```

Verify the SecurityPolicy configuration:

```shell
kubectl get securitypolicy/oidc-example -o yaml
```

## Testing

Port forward gateway 443 port to localhost:

```shell
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')

sudo kubectl -n envoy-gateway-system port-forward service/${ENVOY_SERVICE} 443:443
```

Put www.example.com in the /etc/hosts file in your test machine, so we can use this host name to access the demo from a browser:

```shell
...
127.0.0.1 www.example.com
```

Open a browser and navigate to the `https://www.example.com` address. You should be redirected to the Google login page. After you
successfully login, you should see the response from the backend service.

## Clean-Up

Follow the steps from the [Quickstart](../quickstart) guide to uninstall Envoy Gateway and the example manifest.

Delete the SecurityPolicy and the secret:

```shell
kubectl delete securitypolicy/oidc-example
kubectl delete secret/my-app-client-secret
```

## Next Steps

Checkout the [Developer Guide](../../contributions/develop/) to get involved in the project.

[oidc]: https://openid.net/connect/
[google-oidc]: https://developers.google.com/identity/protocols/oauth2/openid-connect
[SecurityPolicy]: ../../design/security-policy/
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
