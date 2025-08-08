+++
title = "BackendTLSPolicy"
+++


    The `BackendTLSPolicy` resource is Alpha and has been part of
    the Experimental Channel since `v1.0.0`. For more information on release
    channels, refer to our [versioning guide](https://gateway-api.sigs.k8s.io/concepts/versioning).

[BackendTLSPolicy][backendtlspolicy] is a Gateway API type for specifying the TLS configuration
of the connection from the Gateway to a backend pod/s via the Service API object.

## Background

`BackendTLSPolicy` specifically addresses the configuration of TLS in order to convey HTTPS from the Gateway
dataplane to the backend.  This is referred to as "backend TLS termination" and enables the Gateway to know
how to connect to a backend pod that has its own certificate.

While there are other API objects provided for TLS to be configured for **passthrough** and **edge** termination,
this API object allows users to specifically configure **backend** TLS termination.  For more information on TLS
configuration in Gateway API, see [TLS Configuration](https://gateway-api.sigs.k8s.io/guides/tls/).

![Image showing three TLS Termination Types](https://gateway-api.sigs.k8s.io/images/tls-termination-types.png)

BackendTLSPolicy is a Direct [PolicyAttachment](https://gateway-api.sigs.k8s.io/reference/policy-attachment/) without defaults or overrides,
applied to a Service that accesses a backend, where the BackendTLSPolicy resides in the same namespace as the
Service to which it is applied. The BackendTLSPolicy and the Service must reside in the same namespace in order
to prevent the complications involved with sharing trust across namespace boundaries.

All Gateway API Routes that point to a referenced Service should respect a configured BackendTLSPolicy.

## Spec

The specification of a [BackendTLSPolicy][backendtlspolicy] consists of:

- [TargetRefs][targetRefs] - Defines the targeted API object of the policy.  Only Service is allowed.
- [Validation][validation] - Defines the configuration for TLS, including hostname, CACertificateRefs, and
WellKnownCACertificates.
- [Hostname][hostname] - Defines the Server Name Indication (SNI) that the Gateway uses to connect to the backend.
- [CACertificateRefs][caCertificateRefs] - Defines one or more references to objects that contain PEM-encoded TLS certificates,
which are used to establish a TLS handshake between the Gateway and backend Pod.  Either CACertificateRefs or
WellKnownCACertificates may be specified, but not both.
- [WellKnownCACertificates][wellKnownCACertificates] - Specifies whether system CA certificates may be used in the TLS
handshake between the Gateway and backend Pod.  Either CACertificateRefs or WellKnownCACertificates may be specified, but not both.

The following chart outlines the object definitions and relationship:
```mermaid
flowchart LR
    backendTLSPolicy[["<b>backendTLSPolicy</b> <hr><align=left>BackendTLSPolicySpec: spec<br>PolicyStatus: status</align>"]]
    spec[["<b>spec</b><hr>PolicyTargetReferenceWithSectionName: targetRefs <br> BackendTLSPolicyValidation: tls"]]
    status[["<b>status</b><hr>[ ]PolicyAncestorStatus: ancestors"]]
    validation[["<b>tls</b><hr>LocalObjectReference: caCertificateRefs<br>wellKnownCACertificatesType: wellKnownCACertificates/<br>PreciseHostname: hostname"]]
    ancestorStatus[["<b>ancestors</b><hr>AncestorRef: parentReference<br>GatewayController: controllerName<br>[]Condition: conditions"]]
    targetRefs[[<b>targetRefs</b><hr>]]
    service["<b>service</>"]
    backendTLSPolicy -->spec
    backendTLSPolicy -->status
    spec -->targetRefs & validation
    status -->ancestorStatus
    targetRefs -->service
    note[<em>choose only one<hr> caCertificateRefs OR wellKnownCACertificates</em>]
    style note fill:#fff
    validation -.- note
```

The following illustrates a BackendTLSPolicy that configures TLS for a Service serving a backend:
```mermaid
flowchart LR
    client(["client"])
    gateway["Gateway"]
    style gateway fill:#02f,color:#fff
    httproute["HTTP<BR>Route"]
    style httproute fill:#02f,color:#fff
    service["Service"]
    style service fill:#02f,color:#fff
    pod1["Pod"]
    style pod1 fill:#02f,color:#fff
    pod2["Pod"]
    style pod2 fill:#02f,color:#fff
    client -.->|HTTP <br> request| gateway
    gateway --> httproute
    httproute -.->|BackendTLSPolicy|service
    service --> pod1 & pod2
```

### Targeting backends

A BackendTLSPolicy targets a backend Pod (or set of Pods) via one or more TargetRefs to a Service.  This TargetRef is a
required object reference that specifies a Service by its Name, Kind (Service), and optionally its Namespace and Group.
TargetRefs identify the Service/s for which your HTTPRoute requires TLS.


    - Cross-namespace certificate references are not allowed.

### BackendTLSPolicyValidation

A BackendTLSPolicyValidation is the specification for the BackendTLSPolicy and defines the configuration for TLS,
including hostname (for server name indication) and certificates.

#### Hostname

Hostname defines the server name indication (SNI) the Gateway should use in order to connect to the backend, and must
match the certificate served by the backend pod. A hostname is the fully qualified domain name of a network host, as
defined by [RFC 3986][rfc-3986]. Note the following deviations from the “host” part of the URI as defined in the RFC:

- IP addresses are not allowed.

Also note:


    - Wildcard hostnames are not allowed.

#### Certificates

The BackendTLSPolicyValidation must contain a certificate reference of some kind, and contains two ways to configure the
certificate to use for backend TLS, CACertificateRefs and WellKnownCACertificates.  Only one of these may be used per
BackendTLSPolicyValidation.

##### CACertificateRefs

CACertificateRefs refer to one or more PEM-encoded TLS certificates.


    - Cross-namespace certificate references are not allowed.

##### WellKnownCACertificates

If you are working in an environment where specific TLS certificates are not required, and your Gateway API
implementation allows system or default certificates to be used, e.g. in a development environment, you may
set WellKnownCACertificates to "System" to tell the Gateway to use a set of trusted CA Certificates. There may be
some variation in which system certificates are used by each implementation. Refer to documentation from your
implementation of choice for more information.

### PolicyStatus

Status defines the observed state of the BackendTLSPolicy and is not user-configurable.  Check status in the same
way you do for other Gateway API objects to verify correct operation.  Note that the status in BackendTLSPolicy
uses `PolicyAncestorStatus` to allow you to know which parentReference set that particular status.

[backendtlspolicy]: https://gateway-api.sigs.k8s.io/references/spec#gateway.networking.k8s.io/v1alpha3.BackendTLSPolicy
[validation]: https://gateway-api.sigs.k8s.io/references/spec#gateway.networking.k8s.io/v1alpha3.BackendTLSPolicy.Validation
[caCertificateRefs]: https://gateway-api.sigs.k8s.io/references/spec#gateway.networking.k8s.io/v1alpha3.BackendTLSPolicyValidation.CACertificateRefs
[wellKnownCACertificates]: https://gateway-api.sigs.k8s.io/references/spec#gateway.networking.k8s.io/v1alpha3.BackendTLSPolicyValidation.WellKnownCACertificates
[hostname]: https://gateway-api.sigs.k8s.io/references/spec#gateway.networking.k8s.io/v1.PreciseHostname
[rfc-3986]: https://tools.ietf.org/html/rfc3986
[targetRefs]: https://gateway-api.sigs.k8s.io/references/spec#gateway.networking.k8s.io/v1alpha2.PolicyTargetReference
