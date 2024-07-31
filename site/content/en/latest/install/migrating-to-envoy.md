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

1. **Convert Ingress Resources**:

   Given an example Ingress configuration:

   ```yaml
   apiVersion: networking.k8s.io/v1
   kind: Ingress
   metadata:
     name: example-ingress
     namespace: default
   spec:
     ingressClassName: nginx
     rules:
     - host: example.com
       http:
         paths:
         - path: /
           pathType: Prefix
           backend:
             service:
               name: example-service
               port:
                 number: 80
     tls:
     - hosts:
       - example.com
       secretName: example-tls
   ```

   Run the following command to convert your Ingress resources into Gateway API resources:

   ```bash
   ingress2gateway print > gateway-resources.yaml
   ```

   This command will output the converted resources to a file named `gateway-resources.yaml`.

2. **Review and Apply the Resources**:

   Review the `gateway-resources.yaml` file to ensure it meets your requirements. Then, apply it to your Kubernetes cluster:

   ```bash
   kubectl apply -f gateway-resources.yaml
   ```

3. **Verify the Gateway**:

   Verify that the Gateway is correctly set up and working by checking its address:

   ```bash
   kubectl get gateway <gateway-name> -n <namespace> -o jsonpath='{.status.addresses}{"\n"}'
   ```

4. **Update DNS**:

   Update your DNS settings to point to the new Gateway address.

5. **Decommission Old Ingress**:

   Once you have confirmed that traffic is correctly routed through the new Gateway and not through the old Ingress, you can delete the old Ingress resources:

   ```bash
   kubectl delete ingress <ingress-name> -n <namespace>
   ```

This process helps transition your traffic handling from Ingress to Envoy Gateway efficiently, leveraging the `ingress2gateway` tool to automate the conversion.