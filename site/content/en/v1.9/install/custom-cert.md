---
title: Control Plane Authentication Using Custom Certificates
weight: -70
---

Envoy Gateway establishes secure TLS connections for control plane communication between the Envoy Gateway deployment and the Envoy Proxy fleet. By default, the Helm chart generates the required certificates before Envoy Gateway starts.

This guide shows how to create and manage these certificates with cert-manager before installing Envoy Gateway.

## Before you begin

Install [cert-manager](https://cert-manager.io/docs/installation/kubectl/) and [Helm](https://helm.sh/docs/intro/install/) before continuing.

The examples below use the default Kubernetes cluster domain, `cluster.local`. If your cluster uses a different domain, update both the `kubernetesClusterDomain` Helm value and the controller certificate DNS names.

## Configure custom certificates for the control plane

1. Create the namespace where Envoy Gateway and the certificate resources will be installed.

   ```shell
   kubectl create namespace envoy-gateway-system
   ```

2. Set up the CA issuer. This example uses a self-signed issuer to create the root CA.

   **Warning:** Do not use the self-signed issuer in production. Use an issuer backed by a trusted certificate authority.

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

3. Wait for the CA certificate and CA issuer to become ready.

   ```shell
   kubectl wait --for=condition=Ready \
     certificate/envoy-gateway-ca \
     --namespace envoy-gateway-system \
     --timeout=5m

   kubectl wait --for=condition=Ready \
     issuer/eg-issuer \
     --namespace envoy-gateway-system \
     --timeout=5m
   ```

4. Create the certificate for the Envoy Gateway controller. cert-manager stores it in the `envoy-gateway` Secret.

   ```shell
   cat <<EOF | kubectl apply -f -
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

5. Create the certificate for Envoy Proxy. cert-manager stores it in the `envoy` Secret.

   ```shell
   cat <<EOF | kubectl apply -f -
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

6. Create the certificate for the rate-limit service. cert-manager stores it in the `envoy-rate-limit` Secret.

   ```shell
   cat <<EOF | kubectl apply -f -
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

7. Wait for the certificates to become ready.

   ```shell
   kubectl wait --for=condition=Ready \
     certificate/envoy-gateway \
     certificate/envoy \
     certificate/envoy-rate-limit \
     --namespace envoy-gateway-system \
     --timeout=5m
   ```

   Verify the certificate resources and the expected TLS Secrets.

   ```shell
   kubectl get certificates \
     --namespace envoy-gateway-system

   kubectl get secrets \
     envoy-gateway \
     envoy \
     envoy-rate-limit \
     --namespace envoy-gateway-system \
     --output=custom-columns=NAME:.metadata.name,TYPE:.type
   ```

8. Create a Helm values file that specifies the Kubernetes cluster domain used in the controller certificate DNS names.

   ```shell
   cat > custom-cert-values.yaml <<'EOF'
   # Keep this value aligned with the cluster domain used in the
   # envoy-gateway Certificate DNS names.
   kubernetesClusterDomain: cluster.local
   EOF
   ```

   The certificate Secret names are fixed, so no certificate-specific Helm override is required. The chart uses the pre-created Secrets named `envoy-gateway`, `envoy`, and `envoy-rate-limit`.

   Keep the certgen job enabled. It leaves existing certificate Secrets unchanged and creates any additional Secrets required by Envoy Gateway.

9. Install Envoy Gateway using the custom values file.

   ```shell
   helm install eg oci://docker.io/envoyproxy/gateway-helm \
     --version {{< helm-version >}} \
     --namespace envoy-gateway-system \
     --values custom-cert-values.yaml
   ```

10. Wait for Envoy Gateway to become available.

    ```shell
    kubectl wait --for=condition=Available \
      deployment/envoy-gateway \
      --namespace envoy-gateway-system \
      --timeout=5m
    ```

    Verify the deployment, pods, and Helm release.

    ```shell
    kubectl get deployment/envoy-gateway \
      --namespace envoy-gateway-system

    kubectl get pods \
      --namespace envoy-gateway-system

    helm status eg \
      --namespace envoy-gateway-system
    ```
