#!/usr/bin/env bash

set -euo pipefail

# Setup default values
CLUSTER_NAME=${CLUSTER_NAME:-"envoy-gateway"}
METALLB_VERSION=${METALLB_VERSION:-"v0.13.10"}
KIND_NODE_TAG=${KIND_NODE_TAG:-"v1.28.0"}

## Create kind cluster.
if [[ -z "${KIND_NODE_TAG}" ]]; then
  tools/bin/kind create cluster --name "${CLUSTER_NAME}"
else
  tools/bin/kind create cluster --image "kindest/node:${KIND_NODE_TAG}" --name "${CLUSTER_NAME}"
fi

## Install MetalLB.
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/"${METALLB_VERSION}"/config/manifests/metallb-native.yaml
needCreate="$(kubectl get secret -n metallb-system memberlist --no-headers --ignore-not-found -o custom-columns=NAME:.metadata.name)"
if [ -z "$needCreate" ]; then
    kubectl create secret generic -n metallb-system memberlist --from-literal=secretkey="$(openssl rand -base64 128)"
fi

# Wait for MetalLB to become available.
kubectl rollout status -n metallb-system deployment/controller --timeout 5m
kubectl rollout status -n metallb-system daemonset/speaker --timeout 5m

# Apply config with addresses based on docker network IPAM.
subnet=$(docker network inspect kind | jq -r '.[].IPAM.Config[].Subnet | select(contains(":") | not)')
# Assume default kind network subnet prefix of 16, and choose addresses in that range.
address_first_octets=$(echo "${subnet}" | awk -F. '{printf "%s.%s",$1,$2}')
address_range="${address_first_octets}.255.200-${address_first_octets}.255.250"
kubectl apply -f - <<EOF
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  namespace: metallb-system
  name: kube-services
spec:
  addresses:
  - ${address_range}
---
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: kube-services
  namespace: metallb-system
spec:
  ipAddressPools:
  - kube-services
EOF
