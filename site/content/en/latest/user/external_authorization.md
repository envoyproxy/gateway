---
title: "External Authorization"
---

This guide provides instructions for configuring an [External Authorization][ExternalAuthorization] service.  
The external authorization network filter calls an external authorization service to check if the incoming request is authorized or not.

If the external authorization service is not responding a 503 will be returned.  
The current implementation only supports the gRPC service, clear gRPC and TLS are allowed.

## Configuration

The below example defines a [SecurityPolicy][SecurityPolicy] that queries an external authorization server that must respond on `envoy-ext-auth.local` port 10003 using TLS.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: ext-authz-example
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: backend
  extAuthz:
    grpcURI: https://envoy-ext-auth.local:10003

EOF
```

Verify the SecurityPolicy configuration:

```shell
kubectl get securitypolicy/cors-example -o yaml
```

[SecurityPolicy]: ../../design/security-policy/
[ExternalAuthorization]: https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_authz_filter