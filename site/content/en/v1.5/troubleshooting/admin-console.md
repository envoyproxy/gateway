---
title: "Admin Console"
---

Envoy Gateway provides a built-in web-based admin console that offers a comprehensive interface for monitoring, debugging, and managing your Envoy Gateway deployment. The admin console provides real-time visibility into the control plane status, configuration state, performance metrics, and debugging capabilities.

## Prerequisites

{{< boilerplate prerequisites >}}

## Overview

The admin console is automatically enabled and provides the following key features:

- **Dashboard**: Overview of system status and quick access to all features
- **Server Information**: Detailed runtime information about Envoy Gateway components
- **Configuration Dump**: Real-time view of Gateway API resources and their status
- **Statistics**: Control plane metrics in Prometheus format
- **Performance Profiling**: pprof endpoints for debugging and performance analysis

## Accessing the Admin Console

By default, the admin console is available on `localhost:19000`. There are two ways to access it:

### Method 1: Using kubectl port-forward

```shell
kubectl port-forward -n envoy-gateway-system deployment/envoy-gateway 19000:19000
```

Then open your browser and navigate to:

```text
http://localhost:19000
```

### Method 2: Using egctl dashboard command

You can install egctl by following the [installation guide](../install/install-egctl).

```shell
egctl x dashboard eg
```

This command will automatically set up port forwarding and open the admin console in your default browser.

![Admin Console Main Dashboard](/img/admin_main.png)

## Configuration

The admin console can be configured through the `EnvoyGateway` configuration resource:

### Development Configuration

For development environments, you may want to enable additional debugging features:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
metadata:
  name: envoy-gateway-dev
  namespace: envoy-gateway-system
spec:
  admin:
    address:
      # Allow external access (use with caution)
      host: "0.0.0.0"
      port: 19000
    
    # Enable pprof for performance debugging
    enablePprof: true
```

### Production Configuration

For production environments, use more restrictive settings:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
metadata:
  name: envoy-gateway-prod
  namespace: envoy-gateway-system
spec:
  admin:
    address:
      # Localhost only for security
      host: "127.0.0.1"
      port: 19000
    
    # Disable pprof in production
    enablePprof: false
```

## Features

### Dashboard

The main dashboard provides:

- **System Information**: Version, uptime, and platform details
- **Component Status**: Real-time status of all Envoy Gateway components
- **Quick Navigation**: Easy access to all admin console features
- **Auto-refresh**: Automatic updates every 30 seconds

![Admin Console Dashboard](/img/admin_main.png)

### Server Information

The server info page displays:

- **Component Health**: Status of Provider Service, GatewayAPI Translator, xDS Translator, and Infrastructure Manager
- **Runtime Details**: Version information, uptime, and system metrics
- **Configuration**: Current EnvoyGateway configuration settings

![Server Information Page](/img/admin_server_info.png)

### Configuration Dump

The config dump feature provides:

- **Resource Explorer**: Browse all Gateway API resources (Gateways, HTTPRoutes, etc.)
- **Search Functionality**: Find resources by name or namespace
- **Real-time Updates**: Live view of configuration changes
- **JSON Export**: Download complete configuration as JSON

![Configuration Dump Page](/img/admin_config_dump.png)

The search functionality allows you to quickly find specific resources:

![Configuration Search](/img/admin_config_dump_search.png)

You can also view the complete configuration dump in JSON format:

![Complete Configuration Dump](/img/admin_config_dump_all.png)

### Statistics

The statistics page offers:

- **Prometheus Metrics**: All control plane metrics in Prometheus format
- **Metrics Categories**: Organized view of different metric types:
  - Watching Components (event-driven architecture metrics)
  - Status Updater (resource status update metrics)
  - xDS Server (proxy configuration delivery metrics)
  - Infrastructure Manager (Kubernetes resource operation metrics)
  - Wasm (WebAssembly extension metrics)
  - Topology Injector (node topology injection metrics)

Access metrics directly via: `http://localhost:19000/api/metrics`

![Statistics Page](/img/admin_stats.png)

The metrics endpoint provides detailed Prometheus metrics:

![Metrics Endpoint](/img/admin_metrics.png)

### Performance Profiling

When `enablePprof` is set to `true`, the profiling page provides:

- **CPU Profile**: Analyze CPU usage patterns
- **Memory Heap**: Monitor memory allocation and identify leaks
- **Goroutines**: Debug concurrency issues and goroutine leaks
- **Mutex/Block**: Find contention points in the application

![Performance Profiling Page](/img/admin_profiling.png)

{{% alert title="Security Warning" color="warning" %}}
Only enable pprof endpoints in development or debugging scenarios. These endpoints can expose sensitive information and should not be enabled in production environments.
{{% /alert %}}
