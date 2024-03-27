---
title: "Accelerating TLS Handshakes using Private Key Provider in Envoy"
---

TLS operations can be accelerated or the private key can be protected using specialized hardware. This can be leveraged in Envoy using [Envoy Private Key Provider](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/transport_sockets/tls/v3/common.proto#extensions-transport-sockets-tls-v3-privatekeyprovider) is added to Envoy.

Today, there are two private key providers implemented in Envoy as contrib extensions:
- [QAT in Envoy 1.24 release](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/private_key_providers/qat/v3alpha/qat.proto#extensions-private-key-providers-qat-v3alpha-qatprivatekeymethodconfig)
- [CryptoMB in Envoy 1.20 release](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/private_key_providers/cryptomb/v3alpha/cryptomb.proto )

Both of them are used to accelerate the TLS handshake through the hardware capabilities.

This guide will walk you through the steps required to configure TLS Termination mode for TCP traffic while also using the Envoy Private Key Provider to accelerate the TLS handshake by leveraging QAT and the HW accelerator available on Intel SPR/EMR Xeon server platforms.

## Prerequisites

### For QAT

- Install Linux kernel 5.17 or similar
- Ensure the node has QAT devices by checking the QAT physical function devices presented. [Supported Devices](https://intel.github.io/quickassist/qatlib/requirements.html#qat2-0-qatlib-supported-devices)

  ```shell
  echo `(lspci -d 8086:4940 && lspci -d 8086:4941 && lspci -d 8086:4942 && lspci -d 8086:4943 && lspci -d 8086:4946 && lspci -d 8086:4947) | wc -l` supported devices found.
  ```

- Enable IOMMU from BIOS
- Enable IOMMU for Linux kernel

  Figure out the QAT VF device id

  ```shell
  lspci -d 8086:4941 && lspci -d 8086:4943 && lspci -d 8086:4947
  ```

  Attach the QAT device to vfio-pci through kernel parameter by the device id gotten from previous command.

  ```shell
  cat /etc/default/grub:
  GRUB_CMDLINE_LINUX="intel_iommu=on vfio-pci.ids=[QAT device id]"
  update-grub
  reboot
  ````

  Once the system is rebooted, check if the IOMMU has been enabled via the following command:

  ```shell
  dmesg| grep IOMMU
  [    1.528237] DMAR: IOMMU enabled
  ```

- Enable virtual function devices for QAT device

  ```shell
  modprobe vfio_pci
  rmmod qat_4xxx
  modprobe qat_4xxx
  qat_device=$(lspci -D -d :[QAT device id] | awk '{print $1}')
  for i in $qat_device; do echo 16|sudo tee /sys/bus/pci/devices/$i/sriov_numvfs; done
  chmod a+rw /dev/vfio/*
  ```

- Increase the container runtime memory lock limit (using the containerd as example here)

  ```shell
  mkdir /etc/systemd/system/containerd.service.d
  cat <<EOF >>/etc/systemd/system/containerd.service.d/memlock.conf
  [Service]
  LimitMEMLOCK=134217728
  EOF
  ```

  Restart the container runtime (for containerd, CRIO has similar concept)

  ```shell
  systemctl daemon-reload
  systemctl restart containerd
  ```

- Install [Intel® QAT Device Plugin for Kubernetes](https://github.com/intel/intel-device-plugins-for-kubernetes)

  ```shell
  kubectl apply -k 'https://github.com/intel/intel-device-plugins-for-kubernetes/deployments/qat_plugin?ref=main'
  ```

  Verification of the plugin deployment and detection of QAT hardware can be confirmed by examining the resource allocations on the nodes:

  ```shell
  kubectl get node -o yaml| grep qat.intel.com
  ```

### For CryptoMB:

It required the node with 3rd generation Intel Xeon Scalable processor server processors, or later.
- For kubernetes Cluster, if not all nodes that support Intel® AVX-512 in Kubernetes cluster, you need to add some labels to divide these two kinds of nodes manually or using [NFD](https://github.com/kubernetes-sigs/node-feature-discovery).

  ```shell
  kubectl apply -k https://github.com/kubernetes-sigs/node-feature-discovery/deployment/overlays/default?ref=v0.15.1
  ```

- Checking the available nodes with required cpu instructions:
  - Check the node labels if using [NFD](https://github.com/kubernetes-sigs/node-feature-discovery):

    ```shell
    kubectl get nodes -l feature.node.kubernetes.io/cpu-cpuid.AVX512F,feature.node.kubernetes.io/cpu-cpuid.AVX512DQ,feature.node.kubernetes.io/cpu-cpuid.AVX512BW,feature.node.kubernetes.io/cpu-cpuid.AVX512VBMI2,feature.node.kubernetes.io/cpu-cpuid.AVX512IFMA
    ```

  - Check CPUIDS manually on the node if without using NFD:

    ```shell
    cat /proc/cpuinfo |grep avx512f|grep avx512dq|grep avx512bw|grep avx512_vbmi2|grep avx512ifma
    ```

## Installation

* Follow the steps from the [Quickstart Guide](../quickstart) to install Envoy Gateway.

* Lets enable the EnvoyPatchPolicy feature, which will allow us to directly configure the Private Key Provider Envoy Filter, since Envoy Gateway does not directly expose this functionality.

  ```shell
  cat <<EOF | kubectl apply -f -
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: envoy-gateway-config
    namespace: envoy-gateway-system
  data:
    envoy-gateway.yaml: |
      apiVersion: gateway.envoyproxy.io/v1alpha1
      kind: EnvoyGateway
      gateway:
        controllerName: gateway.envoyproxy.io/gatewayclass-controller
      extensionApis:
        enableEnvoyPatchPolicy: true
  EOF
  ```

* After updating the `ConfigMap`, you will need to restart the `envoy-gateway` deployment so the configuration kicks in

  ```shell
  kubectl rollout restart deployment envoy-gateway -n envoy-gateway-system
  ```

## Create gateway for TLS termination

* Follow the instructions in [TLS Termination for TCP](./tls-termination) to setup a TCP gateway to terminate the TLS connection.

* Update GatewayClass for using the envoyproxy image with contrib extensions and requests required resources.

  ```shell
  cat <<EOF | kubectl apply -f -
  apiVersion: gateway.networking.k8s.io/v1
  kind: GatewayClass
  metadata:
    name: eg
  spec:
    controllerName: gateway.envoyproxy.io/gatewayclass-controller
    parametersRef:
      group: gateway.envoyproxy.io
      kind: EnvoyProxy
      name: custom-proxy-config
      namespace: envoy-gateway-system
  EOF
  ```

### Change EnvoyProxy configuration for QAT

Using the envoyproxy image with contrib extensions and add qat resources requesting, ensure the k8s scheduler find out a machine with required resource.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: envoy-gateway-system
spec:
  concurrency: 1
  provider:
    type: Kubernetes
    kubernetes:
      envoyService:
        type: NodePort
      envoyDeployment:
        container:
          image: envoyproxy/envoy-contrib-dev:latest
          resources:
            requests:
              cpu: 1000m
              memory: 4096Mi
              qat.intel.com/cy: '1'
            limits:
              cpu: 1000m
              memory: 4096Mi
              qat.intel.com/cy: '1'
EOF
```

### Change EnvoyProxy configuration for CryptoMB

Using the envoyproxy image with contrib extensions and add the node affinity to scheduling the Envoy Gateway pod on the machine with required CPU instructions.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: envoy-gateway-system
spec:
  concurrency: 1
  provider:
    type: Kubernetes
    kubernetes:
      envoyService:
        type: NodePort
      envoyDeployment:
        container:
          image: envoyproxy/envoy-contrib-dev:latest
          resources:
            requests:
              cpu: 1000m
              memory: 4096Mi
            limits:
              cpu: 1000m
              memory: 4096Mi
        pod:
          affinity:
            nodeAffinity:
              requiredDuringSchedulingIgnoredDuringExecution:
                nodeSelectorTerms:
                - matchExpressions:
                  - key: feature.node.kubernetes.io/cpu-cpuid.AVX512F
                    operator: Exists
                  - key: feature.node.kubernetes.io/cpu-cpuid.AVX512DQ
                    operator: Exists
                  - key: feature.node.kubernetes.io/cpu-cpuid.AVX512BW
                    operator: Exists
                  - key: feature.node.kubernetes.io/cpu-cpuid.AVX512IFMA
                    operator: Exists
                  - key: feature.node.kubernetes.io/cpu-cpuid.AVX512VBMI2
                    operator: Exists
EOF
```

Or using `preferredDuringSchedulingIgnoredDuringExecution` for best effort scheduling, or not doing any node affinity, just doing the random scheduling. The CryptoMB private key provider supports software fallback if the required CPU instructions aren't here.

## Apply EnvoyPatchPolicy to enable private key provider

### Benchmark before enabling private key provider

First follow the instructions in [TLS Termination for TCP](./tls-termination) to do the functionality test.

Ensure the cpu frequency governor set as `performance`.

```shell
export NUM_CPUS=`lscpu | grep "^CPU(s):"|awk '{print $2}'`
for i in `seq 0 1 $NUM_CPUS`; do sudo cpufreq-set -c $i -g performance; done
```

Using the nodeport as the example, fetch the node port from envoy gateway service.

```shell
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
export NODE_PORT=$(kubectl -n envoy-gateway-system get svc/$ENVOY_SERVICE -o jsonpath='{.spec.ports[0].nodePort}')
```

```shell
echo "127.0.0.1 www.example.com" >> /etc/hosts
```

Benchmark the gateway with fortio.

```shell
fortio load -c 10 -k -qps 0 -t 30s -keepalive=false https://www.example.com:${NODE_PORT}
```

### For QAT

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyPatchPolicy
metadata:
  name: key-provider-patch-policy
  namespace: default
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
    namespace: default
  type: JSONPatch
  jsonPatches:
    - type: "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret"
      name: default/example-cert
      operation:
        op: add
        path: "/tls_certificate/private_key_provider"
        value:
          provider_name: qat
          typed_config:
            "@type": "type.googleapis.com/envoy.extensions.private_key_providers.qat.v3alpha.QatPrivateKeyMethodConfig"
            private_key:
              inline_string: |
                abcd
            poll_delay: 0.001s
          fallback: true
    - type: "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret"
      name: default/example-cert
      operation:
        op: copy
        from: "/tls_certificate/private_key"
        path: "/tls_certificate/private_key_provider/typed_config/private_key"
EOF
```

### For CryptoMB

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyPatchPolicy
metadata:
  name: key-provider-patch-policy
  namespace: default
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
    namespace: default
  type: JSONPatch
  jsonPatches:
    - type: "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret"
      name: default/example-cert
      operation:
        op: add
        path: "/tls_certificate/private_key_provider"
        value:
          provider_name: cryptomb
          typed_config:  
            "@type": "type.googleapis.com/envoy.extensions.private_key_providers.cryptomb.v3alpha.CryptoMbPrivateKeyMethodConfig"
            private_key:
              inline_string: |
                abcd
            poll_delay: 0.001s
          fallback: true
    - type: "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret"
      name: default/example-cert
      operation:
        op: copy
        from: "/tls_certificate/private_key"
        path: "/tls_certificate/private_key_provider/typed_config/private_key"
EOF
```

### Benchmark after enabling private key provider

First follow the instructions in [TLS Termination for TCP](./tls-termination) to do the functionality test again.

Benchmark the gateway with fortio.

```shell
fortio load -c 64 -k -qps 0 -t 30s -keepalive=false https://www.example.com:${NODE_PORT}
```

You will see a performance boost after private key provider enabled. For example, you will get results as below.

Without private key provider:

```shell
All done 43069 calls (plus 10 warmup) 6.966 ms avg, 1435.4 qps
```

With CryptoMB private key provider, the QPS is over 2 times than without private key provider.

```shell
All done 93983 calls (plus 128 warmup) 40.880 ms avg, 3130.5 qps
```

With QAT private key provider, the QPS is over 3 times than without private key provider

```shell
All done 134746 calls (plus 128 warmup) 28.505 ms avg, 4489.6 qps
```