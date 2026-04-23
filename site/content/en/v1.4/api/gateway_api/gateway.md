+++
title = "Gateway"
+++


    The `Gateway` resource is GA and has been part of the Standard Channel since
    `v0.5.0`. For more information on release channels, refer to our [versioning
    guide](../concepts/versioning).

A `Gateway` is 1:1 with the lifecycle of the configuration of infrastructure.
When a user creates a `Gateway`, some load balancing infrastructure is
provisioned or configured (see below for details) by the `GatewayClass`
controller. `Gateway` is the resource that triggers actions in this API. Other
resources in this API are configuration snippets until a Gateway has been
created to link the resources together.

The `Gateway` spec defines the following:

*   `GatewayClassName`- Defines the name of a `GatewayClass` object used by
    this Gateway.
*   `Listeners`-  Define the hostnames, ports, protocol, termination, TLS
    settings and which routes can be attached to a listener.
*   `Addresses`- Define the network addresses requested for this gateway.

If the desired configuration specified in Gateway spec cannot be achieved, the
Gateway will be in an error state with details provided by status conditions.

### Deployment models

Depending on the `GatewayClass`, the creation of a `Gateway` could do any of
the following actions:

* Use cloud APIs to create an LB instance.
* Spawn a new instance of a software LB (in this or another cluster).
* Add a configuration stanza to an already instantiated LB to handle the new
  routes.
* Program the SDN to implement the configuration.
* Something else we haven’t thought of yet...

The API does not specify which one of these actions will be taken.

### Gateway Status

`GatewayStatus` is used to surface the status of a `Gateway` relative to the
desired state represented in `spec`. `GatewayStatus` consists of the following:

- `Addresses`- Lists the IP addresses that have actually been bound to the
  Gateway.
- `Listeners`- Provide status for each unique listener defined in `spec`.
- `Conditions`- Describe the current status conditions of the Gateway.

Both `Conditions` and `Listeners.conditions` follow the conditions pattern used
elsewhere in Kubernetes. This is a list that includes a type of condition, the
status of the condition and the last time this condition changed.
