---
title: "Add Pprof support in Envoy Gateway"
---

## Overview

Envoy Gateway exposes endpoints at `localhost:8899/debug/pprof` to run Golang profiles to aid in live debugging. The endpoints are equivalent to those found in the http/pprof package. `/debug/pprof/` returns an HTML page listing the available profiles.

## Goals

* Add Debug Pprof support to Envoy Gateway control plane.
* Define an API to allow Envoy Gateway to custom debug server configuration.

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
* Add `debug` field under `admin` field.
* Add `enable`, `port` and `host` under `address` field.

Here is an example configuration

``` yaml
apiVersion: config.gateway.envoyproxy.io/v1alpha1
gateway:
    controllerName: "gateway.envoyproxy.io/gatewayclass-controller"
kind: EnvoyGateway
provider:
    type: "Kubernetes"
admin:
    debug: true
    address:
        port: 8899
        host: "127.0.0.1"
```
