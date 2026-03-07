---
title: Migrating from Ingress Resources
---

## Introduction

Migrating from Ingress to Envoy Gateway involves converting existing Ingress resources into resources compatible with Envoy Gateway. Two tools are available to help with this migration:

### ingress2gateway

The official `ingress2gateway` tool (maintained by Kubernetes SIG-Network) transforms Ingress resources into Gateway API resources.

### ingress2eg

The `ingress2eg` tool is an **unofficial proof-of-concept** forked from ingress2gateway with additional capabilities:

- **NGINX Annotation Support**: Converts NGINX-specific annotations (16+ feature categories including session affinity, authentication, rate limiting, CORS, canary deployments, etc.)
- **Envoy Gateway CRD Output**: Generates not only Gateway API resources (Gateway, HTTPRoute) but also Envoy Gateway specific CRDs (BackendTrafficPolicy, SecurityPolicy, etc.)

We aim to get this feature merged upstream in `ingress2gateway` as well.

This guide will walk you through the prerequisites, installation of both tools, and provide example migration processes.

## Prerequisites

Before you start the migration, ensure you have the following:

1. **Envoy Gateway Installed**: You need Envoy Gateway set up in your Kubernetes cluster. Follow the [Envoy Gateway installation guide](../install) for details.
2. **Kubernetes Cluster Access**: Ensure you have access to your Kubernetes cluster and necessary permissions to manage resources.
3. **Installation of `ingress2gateway` Tool**: You need to install the `ingress2gateway` tool in your Kubernetes cluster and configure it accordingly. Follow the [ingress2gateway tool installation guide](https://github.com/kubernetes-sigs/ingress2gateway/blob/main/README.md#installation) for details.
4. **Installation of `ingress2eg` Tool**: You need to install the `ingress2eg` tool in your Kubernetes cluster and configure it accordingly. Follow the [ingress2eg tool installation guide](https://github.com/kkk777-7/ingress2eg?tab=readme-ov-file#installation) for details.

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

## Example Migration using ingress2eg

The `ingress2eg` tool provides extended support for NGINX-specific annotations and generates Envoy Gateway CRD resources in addition to Gateway API resources. Follow steps 1-2 from the ingress2gateway example above for Envoy Gateway installation and GatewayClass creation.

### 1. Install ingress2eg

Ensure you have the ingress2eg tool installed. Follow the [installation guide](https://github.com/kkk777-7/ingress2eg?tab=readme-ov-file#installation) for details.

### 2. Prepare Ingress with NGINX Annotations

Here's an example Ingress resource using NGINX annotations for rate limiting and CORS:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: example-ingress-nginx
  namespace: default
  annotations:
    nginx.ingress.kubernetes.io/limit-rps: "10"
    nginx.ingress.kubernetes.io/enable-cors: "true"
    nginx.ingress.kubernetes.io/cors-allow-methods: "GET, POST, OPTIONS"
    nginx.ingress.kubernetes.io/cors-allow-origin: "https://example.com"
spec:
  rules:
  - host: example.com
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: api-service
            port:
              number: 80
```

### 3. Run ingress2eg

Use ingress2eg to convert the Ingress resources. The tool supports various options:

```sh
# Convert from cluster resources in a specific namespace
ingress2eg print --namespace default

# Convert from a file
ingress2eg print --input-file ingress.yaml
```

### 4. Review Generated Resources

For the above Ingress example, ingress2eg generates the following resources:

**Gateway API Resources:**
- Gateway
- HTTPRoute

**Envoy Gateway CRD Resources:**
- **BackendTrafficPolicy**: Generated from rate limiting annotation (`limit-rps`)
- **SecurityPolicy**: Generated from CORS annotations (`enable-cors`, `cors-allow-methods`, `cors-allow-origin`)

Generated BackendTrafficPolicy (from the `limit-rps` annotation):

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: rate-limit-policy
  namespace: default
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: example-httproute
  rateLimit:
    type: Local
    local:
      rules:
      - clientSelectors:
        - sourceCIDR:
            type: Distinct
            value: 0.0.0.0/0
        limit:
          requests: 10
          unit: Second
```

Generated SecurityPolicy (from the CORS annotations):

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: cors-policy
  namespace: default
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: example-httproute
  cors:
    allowOrigins:
    - "https://example.com"
    allowMethods:
    - GET
    - POST
    - OPTIONS
```

### 5. Apply the Generated Resources

Save the output and apply all resources to your cluster:

```sh
ingress2eg print --namespace default > gateway-resources.yaml
kubectl apply -f gateway-resources.yaml
```

### 6. Validate the Migration

Verify that all resources are created correctly:

```sh
kubectl get gateways
kubectl get httproutes
kubectl get backendtrafficpolicies
kubectl get securitypolicies
```

Check the status of the policies to ensure they are accepted:

```sh
kubectl describe backendtrafficpolicy rate-limit-policy
kubectl describe securitypolicy cors-policy
```

## Summary

By following this guide, users can effectively migrate their existing Ingress resources to Envoy Gateway using the Ingress2gateway, Ingress2eg package. Creating a GatewayClass and linking it to the Envoy Gateway controller ensures that the translated resources are properly programmed in the data plane, providing a seamless transition to the Envoy Gateway environment.