---
title: Migrating from Ingress Resources
---

## Introduction

Migrating from Ingress to Envoy Gateway involves converting existing Ingress resources into resources compatible with Envoy Gateway. The `ingress2gateway` tool simplifies this migration by transforming Ingress resources into Gateway API resources that Envoy Gateway can use. This guide will walk you through the prerequisites, installation of the `ingress2gateway` tool, and provide an example migration process.

## Prerequisites

Before you start the migration, ensure you have the following:

1. **Envoy Gateway Installed**: You need Envoy Gateway set up in your Kubernetes cluster. Follow the [Envoy Gateway installation guide](../install) for details.
2. **Kubernetes Cluster Access**: Ensure you have access to your Kubernetes cluster and necessary permissions to manage resources.
3. **Installation of `ingress2gateway` Tool**: You need to install the `ingress2gateway` tool in your Kubernetes cluster and configure it accordingly. Follow the [ingress2gateway tool installation guide](https://github.com/kubernetes-sigs/ingress2gateway/blob/main/README.md#installation) for details.

## Example Migration

Here’s a step-by-step example of migrating from Ingress to Envoy Gateway using `ingress2gateway`:

### 1. Install and Configure Envoy Gateway

Ensure that Envoy Gateway is installed and running in your cluster. Follow the [official Envoy Gateway installation guide](../install) for setup instructions.

### 2. Create a GatewayClass

To ensure the generated HTTPRoutes are programmed correctly in the Envoy Gateway data plane, create a GatewayClass that links to the Envoy Gateway controller.

Create a `GatewayClass` resource:

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: GatewayClass
metadata:
  name: envoy-gateway-class
spec:
  controllerName: gateway.envoyproxy.io/controller
```

Apply this resource:

```sh
kubectl apply -f gatewayclass.yaml
```

### 3. Install Ingress2gateway

Ensure you have the Ingress2gateway package installed. If not, follow the package’s installation instructions.

### 4. Run Ingress2gateway

Use Ingress2gateway to read your existing Ingress resources and translate them into Gateway API resources.

```sh
./ingress2gateway print
```

This command will:
1. Read your Kube config file to extract the cluster credentials and the current active namespace.
2. Search for Ingress and provider-specific resources in that namespace.
3. Convert them to Gateway API resources (Gateways and HTTPRoutes).

#### Example Ingress Configuration

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: example-ingress
  namespace: default
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: example.com
    http:
      paths:
      - path: /foo
        pathType: Prefix
        backend:
          service:
            name: foo-service
            port:
              number: 80
```

### 5. Save the Output

The command will output the equivalent Gateway API resources in YAML/JSON format to stdout. Save this output to a file for further use.

```sh
./ingress2gateway print > gateway-resources.yaml
```

### 6. Apply the Translated Resources

Apply the translated Gateway API resources to your cluster.

```sh
kubectl apply -f gateway-resources.yaml
```

### 7. Create a Gateway Resource

Create a `Gateway` resource specifying the `GatewayClass` created earlier and including the necessary listeners.

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: example-gateway
  namespace: default
spec:
  gatewayClassName: envoy-gateway-class
  listeners:
  - name: http
    protocol: HTTP
    port: 80
    hostname: example.com
```

Apply this resource:

```sh
kubectl apply -f gateway.yaml
```

### 8. Validate the Migration

Ensure the HTTPRoutes and Gateways are correctly set up and that traffic is being routed as expected. Validate the new configuration by checking the status of the Gateway and HTTPRoute resources.

```sh
kubectl get gateways
kubectl get httproutes
```

### 9. Monitor and Troubleshoot

Monitor the Envoy Gateway logs and metrics to ensure everything is functioning correctly. Troubleshoot any issues by reviewing the Gateway and HTTPRoute statuses and Envoy Gateway controller logs.

## Summary

By following this guide, users can effectively migrate their existing Ingress resources to Envoy Gateway using the Ingress2gateway package. Creating a GatewayClass and linking it to the Envoy Gateway controller ensures that the translated resources are properly programmed in the data plane, providing a seamless transition to the Envoy Gateway environment.