---
title: "Debug support in Envoy Gateway"
---

## Overview

Envoy Gateway exposes endpoints at `localhost:19000/debug/pprof` to run Golang profiles to aid in live debugging.

The endpoints are equivalent to those found in the http/pprof package. `/debug/pprof/` returns an HTML page listing the available profiles.

## Goals

* Add admin server to Envoy Gateway control plane, separated with admin server.
* Add pprof support to Envoy Gateway control plane.
* Define an API to allow Envoy Gateway to custom admin server configuration.
* Define an API to allow Envoy Gateway to open envoy gateway config dump in logs.

The following are the different types of profiles end-user can run:

PROFILE	| FUNCTION
-- | --
/debug/pprof/allocs | Returns a sampling of all past memory allocations.
/debug/pprof/block | Returns stack traces of goroutines that led to blocking on synchronization primitives.
/debug/pprof/cmdline | Returns the command line that was invoked by the current program.
/debug/pprof/goroutine | Returns stack traces of all current goroutines.
/debug/pprof/heap | Returns a sampling of memory allocations of live objects.
/debug/pprof/mutex | Returns stack traces of goroutines holding contended mutexes.
/debug/pprof/profile | Returns pprof-formatted cpu profile. You can specify the duration using the seconds GET parameter. The default duration is 30 seconds.
/debug/pprof/symbol | Returns the program counters listed in the request.
/debug/pprof/threadcreate | Returns stack traces that led to creation of new OS threads.
/debug/pprof/trace | Returns the execution trace in binary form. You can specify the duration using the seconds GET parameter. The default duration is 1 second.

## Non Goals

## API

* Add `admin` field in EnvoyGateway config.
* Add `address` field under `admin` field.
* Add `port` and `host` under `address` field.
* Add `enableDumpConfig` field under `admin field.
* Add `enablePprof` field under `admin field.

Here is an example configuration to open admin server and enable Pprof:

``` yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
gateway:
    controllerName: "gateway.envoyproxy.io/gatewayclass-controller"
kind: EnvoyGateway
provider:
    type: "Kubernetes"
admin:
  enablePprof: true
  address:
    host: 127.0.0.1
    port: 19000
```

Here is an example configuration to open envoy gateway config dump in logs:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
gateway:
    controllerName: "gateway.envoyproxy.io/gatewayclass-controller"
kind: EnvoyGateway
provider:
    type: "Kubernetes"
admin:
   enableDumpConfig: true
```
