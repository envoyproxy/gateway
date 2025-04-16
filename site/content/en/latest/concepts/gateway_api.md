---
title: "The Gateway API"
---

## Before You Begin
You may want to be familiar with:
- [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/)

## Overview

The Envoy Gateway API is a set of Kubernetes Custom Resource Definitions (CRDs) created by the Envoy Gateway project to enhance and simplify the use of Envoy Proxy as an API Gateway. These APIs build on top of the Kubernetes Gateway API, adding functionality that is essential for real-world production traffic management—but without exposing users to the full complexity of raw Envoy configuration.

While the Kubernetes Gateway API provides the standard foundation for routing traffic into your cluster using resources like Gateway, HTTPRoute, and TLSRoute, it intentionally leaves many advanced features out of scope. Envoy Gateway fills these gaps.


## Use Cases

Use the Envoy Gateway API when:
- Apply **security policies** (e.g., JWT, OIDC auth) across routes or gateways
- Set **rate limits** or enable **traffic shaping** per client or backend
- Define **mTLS requirements** or configure **TLS passthrough**
- Centralize routing and policy logic using Kubernetes-native resources
- Customize request/response behavior with filters, retries, mirroring, etc.
- Extend your routing logic without writing raw Envoy config

## How the Gateway API fits into Envoy Gateway

Envoy Gateway uses the **Kubernetes Gateway API** as its foundation and layers its own **CRDs** on top to expose advanced features in a user-friendly way.

You define your desired behavior using Kubernetes manifests. For example:

- Use a `SecurityPolicy` to enforce authentication on an `HTTPRoute`.
- Apply a `BackendTrafficPolicy` to configure circuit breaking or retries.
- Attach a `ClientTrafficPolicy` to rate limit based on client identity.

Envoy Gateway (the controller) watches these custom resources and converts them into the correct configuration for **Envoy Proxy** (the data plane). Envoy Proxy then applies these configurations to real-time traffic.

## Related Resources

- [Envoy Gateway API Reference](../api/extension_types)
- [Kubernetes Gateway API Reference](../api/gateway_api/_index.md)
