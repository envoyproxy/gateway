---
title: "Gateway Address"
---

The Gateway API provides an optional [Addresses][] field through which Envoy Gateway can set addresses for Envoy Proxy Service.
Depending on the Service Type, the addresses of gateway can be used as:

- [External IPs](#external-ips)
- [Cluster IP](#cluster-ip)

## Prerequisites

Follow the steps from the [Quickstart](../../quickstart) to install Envoy Gateway and the example manifest.

## External IPs

Using the addresses in `Gateway.Spec.Addresses` as the [External IPs][] of Envoy Proxy Service, 
this will __require__ the address to be of type `IPAddress` and the [ServiceType][] to be of `LoadBalancer` or `NodePort`.

The Envoy Gateway deploys Envoy Proxy Service as `LoadBalancer` by default, 
so you can set the address of the Gateway directly (the address settings here are for reference only):

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

## Cluster IP

Using the addresses in `Gateway.Spec.Addresses` as the [Cluster IP][] of Envoy Proxy Service,
this will __require__ the address to be of type `IPAddress` and the [ServiceType][] to be of `ClusterIP`.


[Addresses]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.GatewayAddress
[External IPs]: https://kubernetes.io/docs/concepts/services-networking/service/#external-ips
[Cluster IP]: https://kubernetes.io/docs/concepts/services-networking/service/#type-clusterip
[ServiceType]: ../../../api/extension_types#servicetype
