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

## Non-Disruptive CA Rotation

When cert-manager rotates the CA certificate (`envoy-gateway-ca`), the leaf cert secrets
(`envoy`, `envoy-gateway`) still contain the old CA in their `ca.crt` field until they are
themselves renewed. This causes xDS connection failures (`CERTIFICATE_VERIFY_FAILED`) for
any proxy pod that reconnects after the old CA expires.

To solve this, use [trust-manager](https://cert-manager.io/docs/trust/trust-manager/) to
distribute the CA independently, and configure Envoy Gateway to read the CA from that
separate path.

### Install trust-manager

trust-manager needs access to the CA secret, so its `trust.namespace` must be set to the
namespace where the CA issuer secret lives (i.e. the namespace where you created the
`envoy-gateway-ca` Certificate above).

```shell
helm install trust-manager jetstack/trust-manager \
  --namespace cert-manager \
  --set app.trust.namespace=envoy-gateway-system
```

### Create a Bundle to sync the CA

```shell
cat <<EOF | kubectl apply -f -
apiVersion: trust.cert-manager.io/v1alpha1
kind: Bundle
metadata:
  name: envoy-gateway-ca-bundle
spec:
  sources:
  - secret:
      name: envoy-gateway-ca
      key: ca.crt
  target:
    configMap:
      key: ca.crt
    namespaceSelector:
      matchLabels:
        kubernetes.io/metadata.name: envoy-gateway-system
EOF
```

When the CA rotates, trust-manager immediately syncs the new CA to the
`envoy-gateway-ca-bundle` ConfigMap.

### Configure Envoy Gateway to use the CA bundle

On the **controller side**, set the CA path and bundle ConfigMap name in your Helm values.
`xdsTLSCABundle` is the name of the ConfigMap created by trust-manager, and it will be
mounted at `/ca-bundle` in the controller pod.

```yaml
xdsTLSCAPath: /ca-bundle/ca.crt
xdsTLSCABundle: envoy-gateway-ca-bundle
```

On the **proxy side**, set `xdsTLSCAPath` in the EnvoyProxy resource and mount the same
ConfigMap:

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
              name: envoy-gateway-ca-bundle
```

With this configuration, when the CA rotates:

1. cert-manager updates the `envoy-gateway-ca` secret
2. trust-manager syncs the new CA to the `envoy-gateway-ca-bundle` ConfigMap
3. Kubernetes propagates the ConfigMap volume update (~60s)
4. The controller picks up the new CA on the next TLS handshake
5. Envoy proxy picks up the new CA via SDS file watch

No pod restarts required.
