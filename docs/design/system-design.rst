System Design
-------------

Goals
~~~~~

-  Define the system components needed to satisfy the requirements of
   Envoy Gateway.

Non-Goals
~~~~~~~~~

-  Create a detailed design and interface specification for each system
   component.

Terminology
~~~~~~~~~~~

-  Control Plane- A collection of inter-related software components for
   providing application gateway and routing functionality. The control
   plane is implemented by Envoy Gateway and provides services for
   managing the data plane. These services are detailed in the
   `components <#components>`__ section.
-  Data Plane- Provides intelligent application-level traffic routing
   and is implemented as one or more Envoy proxies.

Architecture
~~~~~~~~~~~~

.. figure:: ../images/architecture.png
   :alt: Architecture

   Architecture

Configuration
~~~~~~~~~~~~~

Envoy Gateway is configured statically at startup and the managed data
plane is configured dynamically through Kubernetes resources, primarily
`Gateway API <https://gateway-api.sigs.k8s.io>`__ objects.

Static Configuration
^^^^^^^^^^^^^^^^^^^^

Static configuration is used to configure Envoy Gateway at startup,
i.e. change the GatewayClass controllerName, configure a Provider, etc.
Currently, Envoy Gateway only supports configuration through a
configuration file. If the configuration file is not provided, Envoy
Gateway will start up with default configuration parameters.

Dynamic Configuration
^^^^^^^^^^^^^^^^^^^^^

Dynamic configuration is based on the concept of a declaring the desired
state of the data plane and using reconciliation loops to drive the
actual state toward the desired state. The desired state of the data
plane is defined as Kubernetes resources that provide the following
services: \* Infrastructure Management- Manage the data plane
infrastructure, i.e. deploy, upgrade, etc. This configuration is
expressed through
`GatewayClass <https://gateway-api.sigs.k8s.io/concepts/api-overview/#gatewayclass>`__
and
`Gateway <https://gateway-api.sigs.k8s.io/concepts/api-overview/#gateway>`__
resources. A TBD `Custom
Resource <https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/>`__
can be referenced by ``gatewayclass.spec.parametersRef`` to modify data
plane infrastructure default parameters, e.g. expose Envoy network
endpoints using a NodePort service instead of a LoadBalancer service. \*
Traffic Routing- Define how to handle application-level requests to
backend services. For example, route all HTTP requests for
“www.example.com” to a backend service running a web server. This
configuration is expressed through
`HTTPRoute <https://gateway-api.sigs.k8s.io/concepts/api-overview/#httproute>`__
and
`TLSRoute <https://gateway-api.sigs.k8s.io/concepts/api-overview/#tlsroute>`__
resources that match, filter, and route traffic to a
`backend <https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.BackendObjectReference>`__.
Although a backend can be any valid Kubernetes Group/Kind resource,
Envoy Gateway only supports a
`Service <https://kubernetes.io/docs/concepts/services-networking/service/>`__
reference.

Components
~~~~~~~~~~

Envoy Gateway is made up of several components that communicate
in-process; how this communication happens is described in
`watching.md <./watching.md>`__.

Provider
^^^^^^^^

A Provider is an infrastructure component that Envoy Gateway calls to
establish its runtime configuration, resolve services, persist data,
etc. Kubernetes and File are the only supported providers. However,
other providers can be added in the future as Envoy Gateway use cases
are better understood. A provider is configured at start up through
Envoy Gateway’s `static configuration <#static-configuration>`__.

Kubernetes Provider
'''''''''''''''''''

-  Uses Kubernetes-style controllers to reconcile Kubernetes resources
   that comprise the `dynamic configuration <#dynamic-configuration>`__.
-  Manages the data plane through Kubernetes API CRUD operations.
-  Uses Kubernetes for Service discovery.
-  Uses etcd (via Kubernetes API) to persist data.

File Provider
'''''''''''''

-  Uses a file watcher to watch files in a directory that define the
   data plane configuration.
-  Manages the data plane by calling internal APIs,
   e.g. ``CreateDataPlane()``.
-  Uses the host’s DNS for Service discovery.
-  If needed, the local filesystem is used to persist data.

Resource Watcher
^^^^^^^^^^^^^^^^

The Resource Watcher watches resources used to establish and maintain
Envoy Gateway’s dynamic configuration. The mechanics for watching
resources is provider-specific, e.g. informers, caches, etc. are used
for the Kubernetes provider. The Resource Watcher uses the configured
provider for input and provides resources to the Resource Translator as
output.

Resource Translator
^^^^^^^^^^^^^^^^^^^

The Resource Translator translates external resources,
e.g. GatewayClass, from the Resource Watcher to the Intermediate
Representation (IR). It is responsible for: \* Translating
infrastructure-specific resources/fields from the Resource Watcher to
the Infra IR. \* Translating proxy configuration resources/fields from
the Resource Watcher to the xDS IR.

Intermediate Representation (IR)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

The Intermediate Representation defines internal data models that
external resources are translated into. This allows Envoy Gateway to be
decoupled from the external resources used for dynamic configuration.
The IR consists of an Infra IR used as input for the Infra Manager and
an xDS IR used as input for the xDS Translator. \* Infra IR- Used as the
internal definition of the managed data plane infrastructure. \* xDS IR-
Used as the internal definition of the managed data plane xDS
configuration.

xDS Translator
^^^^^^^^^^^^^^

The xDS Translator translates the xDS IR into xDS Resources that are
consumed by the xDS server.

xDS Server
^^^^^^^^^^

The xDS Server is a xDS gRPC Server based on `Go Control
Plane <https://github.com/envoyproxy/go-control-plane>`__. Go Control
Plane implements the xDS Server Protocol and is responsible for using
xDS to configure the data plane.

Infra Manager
^^^^^^^^^^^^^

The Infra Manager is a provider-specific component responsible for
managing the following infrastructure:

-  Data Plane - Manages all the infrastructure required to run the
   managed Envoy proxies. For example, CRUD Deployment, Service, etc.
   resources to run Envoy in a Kubernetes cluster.
-  Auxiliary Control Planes - Optional infrastructure needed to
   implement application Gateway features that require external
   integrations with the managed Envoy proxies. For example, `Global
   Rate
   Limiting <https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/other_features/global_rate_limiting>`__
   requires provisioning and configuring the `Envoy Rate Limit
   Service <https://github.com/envoyproxy/ratelimit>`__ and the `Rate
   Limit
   filter <https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ratelimit/v3/rate_limit.proto#envoy-v3-api-msg-extensions-filters-http-ratelimit-v3-ratelimit>`__.
   Such features are exposed to users through the `Custom Route
   Filters <https://gateway-api.sigs.k8s.io/v1alpha2/api-types/httproute/#filters-optional>`__
   extension.

The Infra Manager consumes the Infra IR as input to manage the data
plane infrastructure.

Design Decisions
~~~~~~~~~~~~~~~~

-  Envoy Gateway will consume one
   `GatewayClass <https://gateway-api.sigs.k8s.io/concepts/api-overview/#gatewayclass>`__
   by comparing its configured controller name with
   ``spec.controllerName`` of a GatewayClass. If multiple GatewayClasses
   exist with the same ``spec.controllerName``, Envoy Gateway will
   follow Gateway API
   `guidelines <https://gateway-api.sigs.k8s.io/concepts/guidelines/#conflicts>`__
   to resolve the conflict. ``gatewayclass.spec.parametersRef`` refers
   to a custom resource for configuring the managed proxy
   infrastructure. If unspecified, default configuration parameters are
   used for the managed proxy infrastructure.
-  Envoy Gateway will manage
   `Gateways <https://gateway-api.sigs.k8s.io/concepts/api-overview/#gateway>`__
   that reference its GatewayClass.

   -  The first Gateway causes Envoy Gateway to provision the managed
      Envoy proxy infrastructure.
   -  Envoy Gateway will merge multiple Gateways that match its
      GatewayClass and will follow Gateway API
      `guidelines <https://gateway-api.sigs.k8s.io/concepts/guidelines/#conflicts>`__
      to resolve any conflicts.
   -  A Gateway ``listener`` corresponds to a proxy
      `Listener <https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listeners#config-listeners>`__.

-  An
   `HTTPRoute <https://gateway-api.sigs.k8s.io/concepts/api-overview/#httproute>`__
   resource corresponds to a proxy
   `Route <https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route.proto#config-route-v3-routeconfiguration>`__.

   -  Each
      `backendRef <https://gateway-api.sigs.k8s.io/v1alpha2/api-types/httproute/#backendrefs-optional>`__
      corresponds to a proxy
      `Cluster <https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto#config-cluster-v3-cluster>`__.

-  The goal is to make the Infra Manager & Translator components
   extensible in the future. For now, extensibility can be achieved
   using xDS support that Envoy Gateway will provide.

The draft for this document is
`here <https://docs.google.com/document/d/1riyTPPYuvNzIhBdrAX8dpfxTmcobWZDSYTTB5NeybuY/edit>`__.

Caveats
~~~~~~~

-  The custom resource used to configure the data plane infrastructure
   is TBD. Track `issue
   95 <https://github.com/envoyproxy/gateway/pull/95>`__ for the latest
   updates.
-  Envoy Gateway’s static configuration spec is currently undefined.
   Track `issue 95 <https://github.com/envoyproxy/gateway/pull/95>`__
   for the latest updates.
