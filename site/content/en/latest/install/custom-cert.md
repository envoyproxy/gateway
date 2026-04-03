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

   ```shell
   cat <<EOF | kubectl apply -f -
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
   EOF
   ```

2. Create a cert for envoy gateway controller, the cert will be stored in secret `envoy-gatewy`.

   ```shell
   cat<<EOF | kubectl apply -f -
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
   EOF
   ```

3. Create a cert for envoy proxy, the cert will be stored in secret `envoy`.

   ```shell
   cat<<EOF | kubectl apply -f -
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
   EOF
   ```

4. Create a cert for rate limit, the cert will be stored in secret `envoy-rate-limit`.

   ```shell
   cat<<EOF | kubectl apply -f -
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
   EOF
   ```

5. Now you can follow the helm chart [installation guide](../install-helm) to install envoy gateway with custom certs.

## Configurable CA Path for xDS mTLS

By default, the CA used for xDS mTLS comes from the `ca.crt` field in the leaf cert
secrets. You can override this on both the controller and proxy to point at a
separately managed CA file.

### Controller side

Set these Helm values to mount a ConfigMap containing your CA and tell the controller
to use it:

```yaml
# Path the controller reads the CA from
xdsTLSCAPath: /ca-bundle/ca.crt
# ConfigMap to mount at /ca-bundle (must exist in the release namespace)
xdsTLSCABundle: my-ca-bundle
```

### Proxy side

Set `xdsTLSCAPath` in the EnvoyProxy resource and mount the CA file:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy
  namespace: envoy-gateway-system
spec:
  xdsTLSCAPath: /ca-bundle/ca.crt
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        container:
          volumeMounts:
          - mountPath: /ca-bundle
            name: ca-bundle
            readOnly: true
        pod:
          volumes:
          - name: ca-bundle
            configMap:
              name: my-ca-bundle
```

The controller re-reads the CA file on each TLS handshake. The proxy picks up
changes via SDS file watch. No pod restarts are required when the CA file is updated.

The CA file supports multiple PEM-encoded certificates. During a CA rotation,
include both the old and new CA in the file so that leaf certs signed by either
CA are accepted. Once all leaf certs have been reissued by the new CA, remove
the old one.
