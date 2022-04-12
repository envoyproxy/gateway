# Introduction

For many application developers, [Envoy][envoy] requires too much cognitive load to support simple proxy
use cases. For example, users must understand concepts such as filters, clusters, etc. to simply proxy
HTTP requests to a personal static website. The goal of Envoy Gateway is to simplify the user experience
while also supporting xDS for more advanced proxy use cases. Envoy Gateway provides:

- The ability for users to run Envoy and proxy traffic to an application with minimal intervention and
  understanding of Envoy, e.g. [xDS][xds].
- An API to support simple proxy use cases while directly supporting xDS to handle advanced proxy use cases.
- Run in Kubernetes and non-Kubernetes environments, e.g. on a Virtual Machine.
- Others?

## Design

Envoy Gateway consists of the following:

- __API Definitions:__ Provides a RESTful API to consumers for managing Envoy and routing traffic between
[upstreams and downstreams][terminology].
- __Control Plane:__ Manages Envoy based on intent described by administration API definitions.
- __Managed Proxy:__ Proxies requests and responses based on intent described by routing API definitions.

![Design Diagram Placeholder](../images/design.png?raw=true "Envoy Gateway Design")

## API Definitions

Provides a RESTful API to consumers for managing Envoy and traffic routing.

## Control Plane

The Control Plane...

## Managed Proxy

Proxies requests and responses based on intent described by routing API definitions.

[envoy]: https://www.envoyproxy.io/
[xds]: https://github.com/envoyproxy/envoy/tree/main/api
[gateway_api]: https://gateway-api.sigs.k8s.io/
[terminology]: https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/intro/terminology
[go_cp]: https://github.com/envoyproxy/go-control-plane
