#!/usr/bin/env bash

set -euo pipefail

# Setup default values
CLUSTER_NAME=${CLUSTER_NAME:-"envoy-gateway"}
METALLB_VERSION=${METALLB_VERSION:-"v0.13.10"}
KIND_NODE_TAG=${KIND_NODE_TAG:-"v1.31.0"}
NUM_WORKERS=${NUM_WORKERS:-""}


KIND_CFG=$(cat <<-EOM
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  ipFamily: dual
nodes:
- role: control-plane
EOM
)

# https://kind.sigs.k8s.io/docs/user/quick-start/#multi-node-clusters
if [[ -n "${NUM_WORKERS}" ]]; then
for _ in $(seq 1 "${NUM_WORKERS}"); do
  KIND_CFG+=$(printf "\n%s" "- role: worker")
done
fi

## Check if kind cluster already exists.
if tools/bin/kind get clusters | grep -q "${CLUSTER_NAME}"; then
  echo "Cluster ${CLUSTER_NAME} already exists."
else
## Create kind cluster.
if [[ -z "${KIND_NODE_TAG}" ]]; then
  cat << EOF | tools/bin/kind create cluster --name "${CLUSTER_NAME}" --config -
${KIND_CFG}
EOF
else
  cat << EOF | tools/bin/kind create cluster --image "kindest/node:${KIND_NODE_TAG}" --name "${CLUSTER_NAME}" --config -
${KIND_CFG}
EOF
fi
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
subnet_v4=$(docker network inspect kind | jq -r '.[].IPAM.Config[] | select(.Subnet | contains(":") | not) | .Subnet')
subnet_v6=$(docker network inspect kind | jq -r '.[].IPAM.Config[] | select(.Subnet | contains(":")) | .Subnet')

# Extract the first three octets for IPv4 (e.g., 192.168.228.0/24 -> 192.168.228)
address_prefix_v4=$(echo "${subnet_v4}" | awk -F. '{print $1"."$2"."$3}')
address_range_v4="${address_prefix_v4}.200-${address_prefix_v4}.250"

# Set IPv6 address range
ipv6_prefix="${subnet_v6%::*}"
address_range_v6="${ipv6_prefix}::200-${ipv6_prefix}::250"

# Print out the address ranges for verification
echo "IPv4 address range: ${address_range_v4}"
echo "IPv6 address range: ${address_range_v6}"

# Apply MetalLB IPAddressPool and L2Advertisement
kubectl apply -f - <<EOF
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  namespace: metallb-system
  name: kube-services
spec:
  addresses:
  - ${address_range_v4}
  - ${address_range_v6}
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
