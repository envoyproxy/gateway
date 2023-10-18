---
title: "Gateway Address"
---

The Gateway API provides an optional [Addresses][] field through which Envoy Gateway can set addresses for Envoy Proxy Service. The currently supported addresses are:

- [External IPs](#External-IPs)

## Installation

Install Envoy Gateway:

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.5.0 -n envoy-gateway-system --create-namespace
```

Wait for Envoy Gateway to become available:

```shell
kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
```

## External IPs

Using the addresses in `Gateway.Spec.Addresses` as the [External IPs][] of Envoy Proxy Service, this will __require__ the address to be of type `IPAddress`.

Install the GatewayClass, Gateway from quickstart:

```shell
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/v0.5.0/quickstart.yaml -n default
```

Set the address of the Gateway, the address settings here are for reference only:

```shell
kubectl patch gateway eg --type=json --patch '[{
   "op": "add",
   "path": "/spec/addresses",
   "value": [{
      "type": "IPAddress",
      "value": "1.2.3.4"
   }]
}]'
```

Verify the Gateway status:

```shell
kubectl get gateway

NAME   CLASS   ADDRESS   PROGRAMMED   AGE
eg     eg      1.2.3.4   True         14m
```

Verify the Envoy Proxy Service status:

```shell
kubectl get service -n envoy-gateway-system

NAME                            TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
envoy-default-eg-64656661       LoadBalancer   10.96.236.219   1.2.3.4       80:31017/TCP   15m
envoy-gateway                   ClusterIP      10.96.192.76    <none>        18000/TCP      15m
envoy-gateway-metrics-service   ClusterIP      10.96.124.73    <none>        8443/TCP       15m
```

__Note:__ If the `Gateway.Spec.Addresses` is explicitly set, it will be the only addresses that populates the Gateway status.

[Addresses]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.GatewayAddress
[External IPs]: https://kubernetes.io/docs/concepts/services-networking/service/#external-ips
