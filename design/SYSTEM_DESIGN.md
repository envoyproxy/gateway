### Architecture
![Architecture](images/architecture.png)

### Configuration

#### User Config
This configuration is based on the [Gateway API](https://gateway-api.sigs.k8s.io) and will provide:
* Infrastructure Management capabilities for the Infrastructure Administrator to provision the infrastructure required to run EnvoyProxy.
This is expressed using the [Gateway resource](https://gateway-api.sigs.k8s.io/concepts/api-overview/#gateway).
* Ingress and API Gateway capabilities for the application developer to define networking and security intent for their incoming traffic.
This is expressed using [HTTPRoute](https://gateway-api.sigs.k8s.io/concepts/api-overview/#httproute), [TLSRoute](https://gateway-api.sigs.k8s.io/concepts/api-overview/#tlsroute),
[TCPRoute or UDPRoute](https://gateway-api.sigs.k8s.io/concepts/api-overview/#tcproute-and-udproute).

#### Bootstrap Config
This is the configuration provided by the Infrastructure Administrator that allows them to bootsrap and configure various internal aspects of EnvoyGateway controller. 
This can either be specified as a commandline argument or be expressed as part of the [GatewayClass resource](https://gateway-api.sigs.k8s.io/concepts/api-overview/#gatewayclass)
that is consumed by a separate process, the EnvoyGateway Operator to create an instance of EnvoyGateway.

### Components

#### Config Sources
This component is responsible for consuming the user configuration from various platforms. Data persistence should be tied to the specific config source’s capabilities. For e.g. in Kubernetes, the resources will persist in etcd, if using the path watcher, the resources will persist in a file.

##### Kubernetes
The Kubernetes controller  watches the Kubernetes API Server for resources, fetches them, and publishes it to the translators for further processing.

##### Path Watcher
It watches for file changes in a path, allowing the user to configure EnvoyGateway using resource configurations saved in a file or directory.

#### Config Server
This is a HTTP/gRPC Server (TBD) allowing EnvoyGateway to be configured from a remote endpoint. 

#### Intermediate Representation (IR)
This is an internal data model that user facing APIs are translated into allowing for internal services & components to be decoupled. 

#### Config Manager
This component consumes the Bootstrap Config, and spawns the appropriate internal services in EnvoyGateway based on the config.

#### Message Service
This component allows internal services to publish message / data types as well as subscribe to them. A message bus architecture allows components to be loosely coupled
, work in an asynchronous manner and also scale out into multiple processes if needed. It can also aggregate resources from multiple publishers allowing configuration from
individual config sources to be aggregated before being processed by the translation layers.

#### Service Resolver
This optional component preprocesses the IR resources and resolves the services into endpoints enabling precise load balancing and resilience policies.
For e.g. in Kubernetes, a controller service could watch for EndpointSlice resources, converting Services to Endpoints, allowing for Envoyproxy to skip kube-proxy’s
load balancing layer. This component is tied to the platform where it is running.  When disabled, the services will be resolved by the underlying DNS resolver or
by explicitly specifying IPs.

#### Gateway API Translator
This is a platform agnostic translator that translates Gateway API resources to an Intermediate Representation.

#### EnvoyProxy Translator
This component translates the IR into EnvoyProxy Resources.

#### xDS Server
This component is a xDS gRPC Server based on the [Envoy Go Control Plane](https://github.com/envoyproxy/go-control-plane) project that implements the xDS Server Protocol
and is responsible for configuring EnvoyProxy resources in EnvoyProxy. 

#### Provisioner
The provisioner configures any infrastruture needed based on the IR.

##### Envoy
Provisions a Envoy based Load balancer service. This is a platform specific component. 
For example, a Terraform or Ansible provisioner could be added in the future to provision the Envoy infra in a non-k8s env.

##### Auxiliary Control Planes
These components are responsible for handling out of band control plane traffic sent by EnvoyProxy.

###### Rate Limit service
This is based on the [Envoy Rate Limit Service](https://github.com/envoyproxy/ratelimit) and will consume the IR and translate it into the server side rate limiting config.
A similar EnvoyProxy translator sub component would translate the IR into Envoy’s ratelimit filter.

### Design Decisions
* A single EnvoyGateway instance will consume many [Gateway resources](https://gateway-api.sigs.k8s.io/concepts/api-overview/#gateway) to manage a fleet of EnvoyProxies with different configurations.
* The goal is to make the Provisioner & Translator layers extensible, but for the near future, extensibility can be achieved using xDS support that EnvoyGateway will provide.

### Open Questions
* Which APIGateway and Ingress features will EnvoyGateway introduce in the near future ?

The draft for this document is [here](https://docs.google.com/document/d/1riyTPPYuvNzIhBdrAX8dpfxTmcobWZDSYTTB5NeybuY/edit)
