---
title: "Goals"
weight: 1
---

The high-level goal of the Envoy Gateway project is to attract more users to Envoy by lowering barriers to adoption
through expressive, extensible, role-oriented APIs that support a multitude of ingress and L7/L4 traffic routing
use cases; and provide a common foundation for vendors to build value-added products without having to re-engineer
fundamental interactions.

## Objectives

### Expressive API

The Envoy Gateway project will expose a simple and expressive API, with defaults set for many capabilities.

The API will be the Kubernetes-native [Gateway API][], plus Envoy-specific extensions and extension points.  This
expressive and familiar API will make Envoy accessible to more users, especially application developers, and make Envoy
a stronger option for "getting started" as compared to other proxies.  Application developers will use the API out of
the box without needing to understand in-depth concepts of Envoy Proxy or use OSS wrappers.  The API will use familiar
nouns that [users](#personas) understand.

The core full-featured Envoy xDS APIs will remain available for those who need more capability and for those who
add functionality on top of Envoy Gateway, such as commercial API gateway products.

This expressive API will not be implemented by Envoy Proxy, but rather an officially supported translation layer
on top.

### Batteries included

Envoy Gateway will simplify how Envoy is deployed and managed, allowing application developers to focus on
delivering core business value.

The project plans to include additional infrastructure components required by users to fulfill their Ingress and API
gateway needs: It will handle Envoy infrastructure provisioning (e.g. Kubernetes Service, Deployment, et cetera), and
possibly infrastructure provisioning of related sidecar services.  It will include sensible defaults with the ability to
override.  It will include channels for improving ops by exposing status through API conditions and Kubernetes status
sub-resources.

Making an application accessible needs to be a trivial task for any developer. Similarly, infrastructure administrators
will enjoy a simplified management model that doesn't require extensive knowledge of the solution's architecture to
operate.

### All environments

Envoy Gateway will support running natively in Kubernetes environments as well as non-Kubernetes deployments.

Initially, Kubernetes will receive the most focus, with the aim of having Envoy Gateway become the de facto
standard for Kubernetes ingress supporting the [Gateway API][].
Additional goals include multi-cluster support and various runtime environments.

### Extensibility

Vendors will have the ability to provide value-added products built on the Envoy Gateway foundation.

It will remain easy for end-users to leverage common Envoy Proxy extension points such as providing an implementation
for authentication methods and rate-limiting.  For advanced use cases, users will have the ability to use the full power
of xDS.

Since a general-purpose API cannot address all use cases, Envoy Gateway will provide additional extension points
for flexibility. As such, Envoy Gateway will form the base of vendor-provided managed control plane solutions,
allowing vendors to shift to a higher management plane layer.

## Non-objectives

### Cannibalize vendor models

Vendors need to have the ability to drive commercial value, so the goal is not to cannibalize any existing vendor
monetization model, though some vendors may be affected by it.

### Disrupt current Envoy usage patterns

Envoy Gateway is purely an additive convenience layer and is not meant to disrupt any usage pattern of any user
with Envoy Proxy, xDS, or go-control-plane.

## Personas

_In order of priority_

### 1. Application developer

The application developer spends the majority of their time developing business logic code.  They require the ability to
manage access to their application.

### 2. Infrastructure administrators

The infrastructure administrators are responsible for the installation, maintenance, and operation of
API gateways appliances in infrastructure, such as CRDs, roles, service accounts, certificates, etc.
Infrastructure administrators support the needs of application developers by managing instances of Envoy Gateway.

[Gateway API]: https://gateway-api.sigs.k8s.io/
