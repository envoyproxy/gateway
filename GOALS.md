# Goals

The high-level goal of the Envoy Gateway project is to attract more users to Envoy by adding new capabilities 
and lowering barriers to adoption.

## Objectives

### Simplified and expressive API
The Envoy Gateway project will introduce a simplified API, with defaults set for many capabilities.

This simplified API will make Envoy accessible to more users, especially application developers, and make Envoy a 
stronger option for "getting started" as compared to Nginx/HAProxy. Application developers will use this simple API
out of the box without needing to understand in-depth concepts of Envoy Proxy or use OSS wrappers. 
The simplified API will use nouns that application developers understand.

The core full-featured Envoy APIs (xDS) will remain available for those who need more capability and for those who 
add functionality on top of Envoy, such as commercial API gateway products.

This simplified API will not be implemented by the Envoy Proxy, but rather an officially supported translation layer 
on top.

### All environments
The Envoy Gateway will support running natively in Kubernetes environments as well as non-Kubernetes deployments.

Initially, Kubernetes will receive the most focus, with the aim of having the Envoy Gateway become the de facto 
standard for Kubernetes ingress supporting the [Gateway API](https://gateway-api.sigs.k8s.io/). 
Medium-term goals include multi-cluster support and various runtime environments.

### Extensibility
Vendors will have the ability to provide value-added products built on the Envoy Gateway foundation.

It will remain easy for end-users to use common Envoy Proxy extension points such as providing an implementation for 
authentication methods and rate-limiting. For advanced use cases, users will have the ability to switch to using xDS 
directly.

Since a general-purpose API cannot address all use cases, the Envoy Gateway will provide additional extension points 
for flexibility. As such, the Envoy Gateway will form the base of vendor-provided managed control plane solutions, 
allowing vendors to shift to a higher management plane layer.

## Non-objectives

### Cannibalize vendor models
Vendors need to have the ability to add value and make money, so the goal is not to cannibalize any existing vendor 
monetization model, though some vendors may be affected by it.

### Disrupt current Envoy usage patterns
The Envoy Gateway is purely an additive convenience layer and is not meant to disrupt any usage pattern of any user 
with Envoy Proxy, xDS, or go-control-plane.

## Personas
_In order of priority_

### 1. Application developer
The application developer spends the majority of their time developing business logic code. They require API gateway 
functionalities to expose their applications. Using expressive configurations, they will define request routes,
TLS termination, rate limits, authentication and authorization policies, etc.

### 2. Infrastructure administrators
The infrastructure administrators are responsible for the installation, maintenance, and operation of
API gateways appliances in infrastructure, such as CRDs, roles, service accounts, certificates, etc.
Infrastructure administrators support the needs of application developers by deploying instances of the Envoy Gateway.

### 3. Envoy developer
The Envoy developer has the ability to quickly develop and test out new or improved features in Envoy proxy, 
that later can be graduated into a user-friendly gateway feature.

## Other

Further discussions and drafts of the project's goals can be found in this document:
https://docs.google.com/document/d/18MuuV9Qzij7Z1OeZ6GrOURKzVi9D0qv2SgvFPELM4gc/edit

