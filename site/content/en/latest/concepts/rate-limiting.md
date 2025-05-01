---
title: "Rate Limiting"
---

## Overview

Rate limiting is a technique used to control the number of requests allowed over a given period to prevent overloading backend services. This helps ensure stable application performance and enhances security by mitigating abuse or potential denial of service.

## Use Cases

Use rate limiting to:

- **Prevent Overload:** Protect backend services from being overwhelmed by excessive traffic.
- **Provide Security:** Mitigate abusive behavior, such as Denial of Service (DoS) attacks.
- **Ensure Fair Usage:** Enforce limits on requests from clients to ensure equitable resource use.

## Rate Limiting in Envoy Gateway

Envoy Gateway leverages Envoy Proxy's powerful rate limiting capabilities through Kubernetes-native configuration. Users define rate limit behavior using the `BackendTrafficPolicy` custom resource, which can be attached to `Gateway`, `HTTPRoute`, or `GRPCRoute` resources. Envoy Gateway supports two primary types of rate limiting:

{{% alert title="Note" color="primary" %}}
Rate limits are applied per route, even if the policy is attached to the Gateway. For example, a limit of 100 requests/second means each route receives its own 100r/s budget.
{{% /alert %}}

### Global Rate Limiting
Global Rate limiting applies limits across multiple Envoy Proxy instances, providing a centralized rate limit that ensures the entire fleet enforces a consistent rate across all proxies. This is useful for distributed environments where you want to enforce traffic limits across the entire system.

To enable global rate limiting, Envoy integrates with an external service that centrally tracks and enforces request limits across all Envoy instances. Envoy provides a [reference implementation](https://github.com/envoyproxy/ratelimit?tab=readme-ov-file#overview) of this service that uses Redis to store request counters. When a request is received, Envoy contacts the service to check whether the request exceeds the allowed rate before continuing.

This model provides:
- **Centralized control:** Uniform enforcement regardless of where the request is processed
- **Multi-tenant protection:** Prevent one user or client from consuming disproportionate resources
- **Upstream shielding:** Reduce load on sensitive systems like databases by throttling excessive requests
- **Consistency during autoscaling:** Global limits aren't tied to the number of replicas

### Local Rate Limiting

Local rate limiting is enforced independently by each instance of Envoy Proxy. Unlike global rate limiting, which coordinates request limits across the entire fleet, local rate limiting evaluates traffic within the scope of a single Envoy instance.

This model is especially useful as a first line of defense against excessive traffic or abuse. Because local rate limiting doesn’t require communication with an external service, it can efficiently block excessive requests before they consume backend resources or overwhelm global limiters.

## Best Practices

- Use local rate limiting when you need simple enforcement with minimal overhead.
- Use global rate limiting for consistency across proxies and to manage shared backend resources.
- Layer local + global rate limits to protect your global rate limit service from bursts.
- Attach rate limit policies to specific routes when applying granular controls.
- Monitor rate limit metrics to fine-tune thresholds based on observed usage patterns.

## Related Resources
- [Task: Global Rate Limit](../tasks/traffic/global-rate-limit.md)
- [Task: Local Rate Limit](../tasks/traffic/local-rate-limit.md)
