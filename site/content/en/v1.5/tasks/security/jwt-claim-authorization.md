---
title: "JWT Claim-Based Authorization"
---

This task provides instructions for configuring JWT claim-based authorization. JWT claim-based authorization checks if an incoming request has the required JWT claims before routing the request to a backend service.

Envoy Gateway introduces a new CRD called [SecurityPolicy][SecurityPolicy] that allows the user to configure JWT claim-based authorization.

This instantiated resource can be linked to a [Gateway][Gateway], [HTTPRoute][HTTPRoute] or [GRPCRoute][GRPCRoute] resource.

## Prerequisites

{{< boilerplate prerequisites >}}

## Configuration

### Create a SecurityPolicy

Please note that the JWT claim-based authorization requires the JWT token to be present in the request. A JWT authentication must be configured in the same SecurityPolicy to validate the JWT token and extract the claims.

The below SecurityPolicy configuration allows requests with a valid JWT token that has the following claims:
- `user.name` claim with the value `John Doe`
- `user.roles` claim with the value `admin`
- `scope` claim with the values `read`, `add`, and `modify`

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: authorization-jwt-claim
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  jwt:
    providers:
    - name: example
      issuer: https://foo.bar.com
      remoteJWKS:
        uri: https://raw.githubusercontent.com/envoyproxy/gateway/refs/heads/main/examples/kubernetes/jwt/jwks.json
  authorization:
    defaultAction: Deny
    rules:
    - name: "allow"
      action: Allow
      principal:
        jwt:
          provider: example
          scopes: ["read", "add", "modify"]
          claims:
          - name: user.name
            values: ["John Doe"]
          - name: user.roles
            valueType: StringArray
            values: ["admin"]
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
  name: authorization-jwt-claim
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  jwt:
    providers:
    - name: example
      issuer: https://foo.bar.com
      remoteJWKS:
        uri: https://raw.githubusercontent.com/envoyproxy/gateway/refs/heads/main/examples/kubernetes/jwt/jwks.json
  authorization:
    defaultAction: Deny
    rules:
    - name: "allow"
      action: Allow
      principal:
        jwt:
          provider: example
          scopes: ["read", "add", "modify"]
          claims:
          - name: user.name
            values: ["John Doe"]
          - name: user.roles
            valueType: StringArray
            values: ["admin"]
```

{{% /tab %}}
{{< /tabpane >}}

Verify the SecurityPolicy configuration:

```shell
kubectl get securitypolicy/authorization-jwt-claim -o yaml
```

## Testing

Ensure the `GATEWAY_HOST` environment variable from the [Quickstart](../../quickstart) is set. If not, follow the
Quickstart instructions to set the variable.

```shell
echo $GATEWAY_HOST
```

Define a JWT token with the required claims.

```shell
export VALID_TOKEN="eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6ImI1MjBiM2MyYzRiZDc1YTEwZTljZWJjOTU3NjkzM2RjIn0.eyJpc3MiOiJodHRwczovL2Zvby5iYXIuY29tIiwic3ViIjoiMTIzNDU2Nzg5MCIsInVzZXIiOnsibmFtZSI6IkpvaG4gRG9lIiwiZW1haWwiOiJqb2huLmRvZUBleGFtcGxlLmNvbSIsInJvbGVzIjpbImFkbWluIiwiZWRpdG9yIl19LCJwcmVtaXVtX3VzZXIiOnRydWUsImlhdCI6MTUxNjIzOTAyMiwic2NvcGUiOiJyZWFkIGFkZCBkZWxldGUgbW9kaWZ5In0.P36iAlmiRCC79OiB3vstF5Q_9OqUYAMGF3a3H492GlojbV6DcuOz8YIEYGsRSWc-BNJaBKlyvUKsKsGVPtYbbF8ajwZTs64wyO-zhd2R8riPkg_HsW7iwGswV12f5iVRpfQ4AG2owmdOToIaoch0aym89He1ZzEjcShr9olgqlAbbmhnk-namd1rP-xpzPnWhhIVI3mCz5hYYgDTMcM7qbokM5FzFttTRXAn5_Luor23U1062Ct_K53QArwxBvwJ-QYiqcBycHf-hh6sMx_941cUswrZucCpa-EwA3piATf9PKAyeeWHfHV9X-y8ipGOFg3mYMMVBuUZ1lBkJCik9f9kboRY6QzpOISARQj9PKMXfxZdIPNuGmA7msSNAXQgqkvbx04jMwb9U7eCEdGZztH4C8LhlRjgj0ZdD7eNbRjeH2F6zrWyMUpGWaWyq6rMuP98W2DWM5ZflK6qvT1c7FuFsWPvWLkgxQwTWQKrHdKwdbsu32Sj8VtUBJ0-ddEb"
```

Decode the JWT token to verify that it has the required claims.

```shell
jq -R 'split(".") | .[0],.[1] | @base64d | fromjson' <<< $(echo ${VALID_TOKEN})
```

The decoded JWT token should look like the following:

```json
{
  "typ": "JWT",
  "alg": "RS256",
  "kid": "b520b3c2c4bd75a10e9cebc9576933dc"
}
{
  "iss": "https://foo.bar.com",
  "sub": "1234567890",
  "user": {
    "name": "John Doe",
    "email": "john.doe@example.com",
    "roles": [
      "admin",
      "editor"
    ]
  },
  "premium_user": true,
  "iat": 1516239022,
  "scope": "read add delete modify"
}
```

Send a request to the backend service with the valid JWT token:

```shell
curl  -H "Host: www.example.com" -H "Authorization: Bearer ${VALID_TOKEN}" "http://${GATEWAY_HOST}/"
```

The request should be allowed and you should see the response from the backend service.

Define a JWT token without the required claims.

```shell
export INVALID_TOKEN="eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6ImI1MjBiM2MyYzRiZDc1YTEwZTljZWJjOTU3NjkzM2RjIn0.eyJpc3MiOiJodHRwczovL2Zvby5iYXIuY29tIiwic3ViIjoiMTIzNDU2Nzg5MCIsInVzZXIiOnsibmFtZSI6IkFsaWNlIFNtaXRoIiwiZW1haWwiOiJhbGljZS5zbWl0aEBleGFtcGxlLmNvbSIsInJvbGVzIjpbImRldmVsb3BlciJdfSwicHJlbWl1bV91c2VyIjpmYWxzZSwiaWF0IjoxNTE2MjM5MDIyLCJzY29wZSI6InJlYWQgYWRkIGRlbGV0ZSJ9.Da547nNXzuQXm5E7LuLAiyFswXsW4RDhuitD_rpadtR7PTwzzOsJoqrVWJ_u1jJDaOTWIpLF4gwxDoY-Aoz_couzXzlAbECLs45ZFoc_UdffpfIbGKqTZx8VtwKuDLFsAeDDDqqx1flxFhvXHftJJdZYr1FgFz9u-absMmRU90DLmEZX3Hnyc8k8eBgeiu6vsWUD0-aNy8cWkFRbwRggkGmucFyUTG8Z1MY3iyH5E66W-ISoX8G9bzE9PTxVAAPDTvefD5iLJPSDJ8qV69OuMCJ8Dczq0L9Dd_w0sF-D1s9MTvexmGg4zBWluJ3r-pU9NHEdhqBypehp_yH8xF5Rt9AE7stZ4oPFZNyfrtkE-4IOnSEkMmzcC65g_rscn0ycerv4N5ZNpkr0x2IYYM4iGuo-ULv5Htnli3rffST45kx1XA8cdsrT1D0K3aPxdIxDIk8sTJf5-WVqRyo-bwxXXltwQLB9jCM_7QbTWQBYAJwUpi-0RW4jCl44-42gZnXf"
```

Decode the JWT token to verify that it does not have the required claims.

```shell
jq -R 'split(".") | .[0],.[1] | @base64d | fromjson' <<< $(echo ${INVALID_TOKEN})
```

The decoded JWT token should look like the following:

```json
{
  "typ": "JWT",
  "alg": "RS256",
  "kid": "b520b3c2c4bd75a10e9cebc9576933dc"
}
{
  "iss": "https://foo.bar.com",
  "sub": "1234567890",
  "user": {
    "name": "Alice Smith",
    "email": "alice.smith@example.com",
    "roles": [
      "developer"
    ]
  },
  "premium_user": false,
  "iat": 1516239022,
  "scope": "read add delete"
}
```

Send a request to the backend service with the invalid JWT token:

```shell
curl  -v -H "Host: www.example.com" -H "Authorization: Bearer ${INVALID_TOKEN}" "http://${GATEWAY_HOST}/"
```

The request should be denied and you should see a `403 Forbidden` response.

## Clean-Up

Follow the steps from the [Quickstart](../../quickstart) to uninstall Envoy Gateway and the example manifest.

Delete the SecurityPolicy and the ClientTrafficPolicy

```shell
kubectl delete securitypolicy/authorization-jwt-claim
```

## Next Steps

Checkout the [Developer Guide](../../../contributions/develop) to get involved in the project.

[SecurityPolicy]: ../../../contributions/design/security-policy
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute
