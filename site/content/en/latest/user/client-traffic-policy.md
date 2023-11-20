---
title: "Client Traffic Policy"
---

This guide explains the usage of the [ClientTrafficPolicy][] API.


## Introduction

The [ClientTrafficPolicy][] API allows system administrators to configure
the behavior for how the Envoy Proxy server behaves with downstream clients.

## Motivation

This API was added as a new policy attachment resource that can be applied to Gateway resources and it is meant to hold settings for configuring behavior of the connection between the downstream client and Envoy Proxy listener. It the counterpart to the [BackendTrafficPolicy][] API resource.

## Quickstart

### Prerequisites

* Follow the steps from the [Quickstart](../quickstart) guide to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

### Enable ClientTrafficPolicy

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: enable-proxy-protocol-policy
  namespace: default
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
    namespace: default
  enableProxyProtocol: true
EOF
```
