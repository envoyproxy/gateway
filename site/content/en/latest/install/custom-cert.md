---
title: Control Plane Authentication using custom certs
weight: -70
---

Envoy Gateway establishes a secure TLS connection for control plane communication between Envoy Gateway pods and the Envoy Proxy fleet. The TLS Certificates used here are self signed and generated using a job that runs before envoy gateway is created, and these certs and mounted on to the envoy gateway and envoy proxy pods. 

This task will walk you through configuring custom certs for control plane auth.

## Before you begin

We use Cert-Manager to manage the certificates. You can install it by following the [official guide](https://cert-manager.io/docs/installation/kubernetes/).

## Configure custom certs for control plane

1. First you need to set up the CA issuer, in this task, we use the `selfsigned-issuer` as an example.

   *You should not use the self-signed issuer in production, you should use a real CA issuer.*

   ```yaml
   ---
   apiVersion: cert-manager.io/v1
   kind: Issuer
   metadata:
     labels:
       app.kubernetes.io/name: envoy-gateway
     name: selfsigned-issuer
     namespace: envoy-gateway-system
   spec:
     selfSigned: {}
   ---
   apiVersion: cert-manager.io/v1
   kind: Certificate
   metadata:
     name: envoy-gateway-ca
     namespace: envoy-gateway-system
   spec:
     isCA: true
     commonName: envoy-gateway
     secretName: envoy-gateway-ca
     privateKey:
       algorithm: RSA
       size: 2048
     issuerRef:
       name: selfsigned-issuer
       kind: Issuer
       group: cert-manager.io
   ---
   apiVersion: cert-manager.io/v1
   kind: Issuer
   metadata:
     labels:
       app.kubernetes.io/name: envoy-gateway
     name: eg-issuer
     namespace: envoy-gateway-system
   spec:
     ca:
       secretName: envoy-gateway-ca
   ```

2. Create a cert for envoy gateway controller, the cert will be stored in secret `envoy-gateway`.

   ```yaml
   ---
   apiVersion: cert-manager.io/v1
   kind: Certificate
   metadata:
     labels:
       app.kubernetes.io/name: envoy-gateway
     name: envoy-gateway
     namespace: envoy-gateway-system
   spec:
     commonName: envoy-gateway
     dnsNames:
     - "envoy-gateway"
     - "envoy-gateway.envoy-gateway-system"
     - "envoy-gateway.envoy-gateway-system.svc"
     - "envoy-gateway.envoy-gateway-system.svc.cluster.local"
     issuerRef:
       kind: Issuer
       name: eg-issuer
     usages:
     - "digital signature"
     - "data encipherment"
     - "key encipherment"
     - "content commitment"
     secretName: envoy-gateway
   ```

3. Create a cert for envoy proxy, the cert will be stored in secret `envoy`.

   ```yaml
   ---
   apiVersion: cert-manager.io/v1
   kind: Certificate
   metadata:
     labels:
       app.kubernetes.io/name: envoy-gateway
     name: envoy
     namespace: envoy-gateway-system
   spec:
     commonName: "*"
     dnsNames:
     - "*.envoy-gateway-system"
     issuerRef:
       kind: Issuer
       name: eg-issuer
     usages:
     - "digital signature"
     - "data encipherment"
     - "key encipherment"
     - "content commitment"
     secretName: envoy
   ```

4. Create a cert for rate limit, the cert will be stored in secret `envoy-rate-limit`.

   ```yaml
   ---
   apiVersion: cert-manager.io/v1
   kind: Certificate
   metadata:
     labels:
       app.kubernetes.io/name: envoy-gateway
     name: envoy-rate-limit
     namespace: envoy-gateway-system
   spec:
     commonName: "*"
     dnsNames:
     - "*.envoy-gateway-system"
     issuerRef:
       kind: Issuer
       name: eg-issuer
     usages:
     - "digital signature"
     - "data encipherment"
     - "key encipherment"
     - "content commitment"
     secretName: envoy-rate-limit
   ```

5. Now you can follow the helm chart [installation guide](../install-helm) to install envoy gateway with custom certs.
