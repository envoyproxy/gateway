---
title: 使用自定义证书的控制平面身份验证
weight: -70
---

Envoy Gateway 为 Envoy Gateway Pod 和 Envoy 代理队列之间的控制平面通信建立了安全的 TLS 连接。
此处使用的 TLS 证书是自签名的，并使用在创建 Envoy Gateway 之前运行的 Job 生成，
并且这些证书被安装到 Envoy Gateway 和 Envoy 代理 Pod 上。

此任务将引导您完成为控制平面身份验证配置自定义证书。

## 开始之前 {#before-you-begin}

我们使用 Cert-Manager 来管理证书。
您可以按照[官方指南](https://cert-manager.io/docs/installation/kubernetes/)安装它。

## 为控制平面配置自定义证书 {#configure-custom-certs-for-control-plane}

1. 首先您需要设置 CA 颁发者，在此任务中，我们以 `selfsigned-issuer` 为例。

   **您不应在生产中使用自签名颁发者，您应该使用真实的 CA 颁发者。**

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

2. 为 Envoy Gateway 控制器创建一个证书，该证书将存储在 `envoy-gatewy` Secret 中。

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

3. 为 Envoy 代理创建一个证书，该证书将存储在 `envoy` Secret 中。

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

4. 创建限流证书，该证书将存储在 `envoy-rate-limit` Secret 中。

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

5. 现在您可以按照 helm Chart [安装指南](../install-helm)使用自定义证书安装 Envoy Gateway。
