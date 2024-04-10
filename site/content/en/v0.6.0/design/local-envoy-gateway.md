---
title: "Running Envoy Gateway locally"
---

## Overview

Today, Envoy Gateway runs only on Kubernetes. This is an ideal solution
when the applications are running in Kubernetes.
However there might be cases when the applications are running on the host which would
require Envoy Gateway to run locally.

## Goals

* Define an API to allow Envoy Gateway to retrieve configuration while running locally.
* Define an API to allow Envoy Gateway to deploy the managed Envoy Proxy fleet on the host
machine.

## Non Goals

* Support multiple ways to retrieve configuration while running locally.
* Support multiple ways to deploy the Envoy Proxy fleet locally on the host.

## API

* The `provider` field within the `EnvoyGateway` configuration only supports
`Kubernetes` today which provides two features - the ability to retrieve
resources from the Kubernetes API Server as well as deploy the managed
Envoy Proxy fleet on Kubernetes. 
* This document proposes adding a new top level `provider` type called `Custom` 
with two fields called `resource` and `infrastructure` to allow the user to configure
the sub providers for providing resource configuration and an infrastructure to deploy
the Envoy Proxy data plane in.
* A `File` resource provider will be introduced to enable retrieving configuration locally
by reading from the configuration from a file.
* A `Host` infrastructure provider will be introduced to allow Envoy Gateway to spawn a 
Envoy Proxy child process on the host.

Here is an example configuration

```
provider:
  type: Custom
  custom:
    resource:
      type: File
      file:
        paths: 
        - "config.yaml"
    infrastructure:
      type: Host
      host: {}
```
