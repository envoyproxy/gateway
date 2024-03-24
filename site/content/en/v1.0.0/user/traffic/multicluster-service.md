---
title: "Multicluster Service Routing"
---

The Multicluster Service API ServiceImport object can be used as part of the GatewayAPI backendRef for configuring routes. For more information about multicluster service API follow [sig documentation](https://multicluster.sigs.k8s.io/concepts/multicluster-services-api/).

We will use [Submariner project](https://github.com/submariner-io/submariner) for setting up the multicluster environment for exporting the service to be routed from peer clusters.

## Setting KIND clusters and installing Submariner.

- We will be using KIND clusters to demonstrate this example.

```shell
git clone https://github.com/submariner-io/submariner-operator
cd submariner-operator
make clusters
```

Note: remain in submariner-operator directory for the rest of the steps in this section

- Install subctl:

```shell
curl -Ls https://get.submariner.io  | VERSION=v0.14.6 bash
```

- Set up multicluster service API and submariner for cross cluster traffic using ServiceImport

```shell
subctl deploy-broker --kubeconfig output/kubeconfigs/kind-config-cluster1 --globalnet
subctl join --kubeconfig output/kubeconfigs/kind-config-cluster1 broker-info.subm --clusterid cluster1 --natt=false
subctl join --kubeconfig output/kubeconfigs/kind-config-cluster2 broker-info.subm --clusterid cluster2 --natt=false
```

Once the above steps are done and all the pods are up in both the clusters. We are ready for installing envoy gateway.

## Install EnvoyGateway

Install the Gateway API CRDs and Envoy Gateway in cluster1:

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v1.0.0 -n envoy-gateway-system --create-namespace --kubeconfig output/kubeconfigs/kind-config-cluster1
```

Wait for Envoy Gateway to become available:

```shell
kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available --kubeconfig output/kubeconfigs/kind-config-cluster1
```

## Install Application

Install the backend application in cluster2 and export it through subctl command.

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/application.yaml --kubeconfig output/kubeconfigs/kind-config-cluster2
subctl export service backend --namespace default --kubeconfig output/kubeconfigs/kind-config-cluster2
```

## Create Gateway API Objects

Create the Gateway API objects GatewayClass, Gateway and HTTPRoute in cluster1 to set up the routing.

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/multicluster-service.yaml --kubeconfig output/kubeconfigs/kind-config-cluster1
```

## Testing the Configuration

Get the name of the Envoy service created the by the example Gateway:

```shell
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
```

Port forward to the Envoy service:

```shell
kubectl -n envoy-gateway-system port-forward service/${ENVOY_SERVICE} 8888:80 &
```

Curl the example app through Envoy proxy:

```shell
curl --verbose --header "Host: www.example.com" http://localhost:8888/get
```
