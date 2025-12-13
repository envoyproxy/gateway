---
title: "Quickstart"
weight: 1
description: Get started with Envoy Gateway in a few simple steps.
---

# Quickstart

Welcome! This guide will help you get Envoy Gateway up and running in a few minutes.

If you’re new to Envoy Gateway, take a moment to understand why it exists and what it is.


## Why Envoy Gateway?

Modern teams need a simpler way to manage how traffic flows into their applications without sacrificing security, reliability, or visibility. That’s why **Envoy Gateway** exists: to make powerful gateway capabilities accessible to both developers and platform engineers.

It achieves this by building on the proven foundation of [Envoy Proxy](https://www.envoyproxy.io/) and packaging it into a **Kubernetes-native API gateway** experience. Envoy Gateway integrates cleanly with Kubernetes through standard resources like `Gateway`, `HTTPRoute`, and `GatewayClass`, so teams can configure traffic and security policies declaratively—no complex Envoy configuration required.

## What Is Envoy Gateway?

At its core, Envoy Gateway is an open-source project that uses Envoy Proxy to implement the Kubernetes Gateway API. It acts as a full-featured, Kubernetes-native API gateway, providing out-of-the-box support for essential functions such as:

- **Traffic Management**: Advanced routing, traffic splitting, and retries.
- **Security**: TLS termination, authentication, and authorization policies.
- **Observability**: Rich metrics, logs, and traces for monitoring traffic.
- **Extensibility**: Support for custom filters to inject your own logic.

In short, it’s simple for developers, powerful for platform engineers, and ready for production at scale.


## Prerequisites

A Kubernetes cluster.

__Note:__ Refer to the [Compatibility Matrix](/news/releases/matrix) for supported Kubernetes versions.

__Note:__ In case your Kubernetes cluster does not have a LoadBalancer implementation, we recommend installing one
so the `Gateway` resource has an Address associated with it. We recommend using [MetalLB](https://metallb.universe.tf/installation/).

## Installation

Install the Gateway API CRDs and Envoy Gateway:

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version {{< helm-version >}} -n envoy-gateway-system --create-namespace
```

Wait for Envoy Gateway to become available:

```shell
kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
```

Install the [GatewayClass][], [Gateway][], [HTTPRoute][] and example app:

```shell
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/{{< yaml-version >}}/quickstart.yaml -n default
```

**Note**: [`quickstart.yaml`] defines that Envoy Gateway will listen for
traffic on port 80 on its globally-routable IP address, to make it easy to use
browsers to test Envoy Gateway. When Envoy Gateway sees that its Listener is
using a privileged port (<1024), it will map this internally to an
unprivileged port, so that Envoy Gateway doesn't need additional privileges.
It's important to be aware of this mapping, since you may need to take it into
consideration when debugging.

[`quickstart.yaml`]: https://github.com/envoyproxy/gateway/releases/download/{{< yaml-version >}}/quickstart.yaml

## Testing the Configuration

{{< tabpane text=true >}}
{{% tab header="With External LoadBalancer Support" %}}

You can also test the same functionality by sending traffic to the External IP. To get the external IP of the
Envoy service, run:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

In certain environments, the load balancer may be exposed using a hostname, instead of an IP address. If so, replace
`ip` in the above command with `hostname`.

Curl the example app through Envoy proxy:

```shell
curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST/get
```

{{% /tab %}}
{{% tab header="Without LoadBalancer Support" %}}

Get the name of the Envoy service created the by the example Gateway:

```shell
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
```

Port forward to the Envoy service:

```shell
kubectl -n envoy-gateway-system port-forward service/${ENVOY_SERVICE} 8888:80 &
```

Curl the example app through Envoy proxy:

```shell
curl --verbose --header "Host: www.example.com" http://localhost:8888/get
```

{{% /tab %}}
{{< /tabpane >}}

## What to explore next?

In this quickstart, you have:
- Installed Envoy Gateway
- Deployed a backend service, and a gateway
- Configured the gateway using Kubernetes Gateway API resources [Gateway][] and [HTTPRoute][] to direct incoming requests over HTTP to the backend service.

Here is a suggested list of follow-on tasks to guide you in your exploration of Envoy Gateway:

- [HTTP Routing](traffic/http-routing)
- [Traffic Splitting](traffic/http-traffic-splitting)
- [Secure Gateways](security/secure-gateways/)
- [Global Rate Limit](traffic/global-rate-limit/)
- [gRPC Routing](traffic/grpc-routing/)

Review the [Tasks](./) section for the scenario matching your use case.  The Envoy Gateway tasks are organized by category: traffic management, security, extensibility, observability, and operations.

## Clean-Up

Use the steps in this section to uninstall everything from the quickstart.

Delete the [GatewayClass][], [Gateway][], [HTTPRoute][] and Example App:

```shell
kubectl delete -f https://github.com/envoyproxy/gateway/releases/download/{{< yaml-version >}}/quickstart.yaml --ignore-not-found=true
```

Delete the Gateway API CRDs and Envoy Gateway:

```shell
helm uninstall eg -n envoy-gateway-system
```

[GatewayClass]: https://gateway-api.sigs.k8s.io/api-types/gatewayclass/
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
