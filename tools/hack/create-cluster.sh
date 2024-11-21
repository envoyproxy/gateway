#!/usr/bin/env bash

set -euo pipefail

# Setup default values
CLUSTER_NAME=${CLUSTER_NAME:-"envoy-gateway"}
METALLB_VERSION=${METALLB_VERSION:-"v0.13.10"}
KIND_NODE_TAG=${KIND_NODE_TAG:-"v1.31.0"}
NUM_WORKERS=${NUM_WORKERS:-""}
IP_FAMILY=${IP_FAMILY:-"ipv4"}

KIND_CFG=$(cat <<-EOM
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  ipFamily: ${IP_FAMILY}
  # it's to prevent inherit search domains from the host which slows down DNS resolution
  # and cause problems to IPv6 only clusters running on IPv4 host.
  dnsSearch: []
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
address_ranges=""

if [ "${IP_FAMILY}" = "ipv4" ] || [ "${IP_FAMILY}" = "dual" ]; then
    subnet_v4=$(docker network inspect kind | jq -r '.[].IPAM.Config[] | select(.Subnet | contains(":") | not) | .Subnet')
    address_prefix_v4=$(echo "${subnet_v4}" | awk -F. '{print $1"."$2"."$3}')
    address_range_v4="${address_prefix_v4}.200-${address_prefix_v4}.250"
    echo "IPv4 address range: ${address_range_v4}"
    address_ranges+="- ${address_range_v4}"
fi

if [ "${IP_FAMILY}" = "ipv6" ] || [ "${IP_FAMILY}" = "dual" ]; then
    subnet_v6=$(docker network inspect kind | jq -r '.[].IPAM.Config[] | select(.Subnet | contains(":")) | .Subnet')
    ipv6_prefix="${subnet_v6%::*}"
    address_range_v6="${ipv6_prefix}::c8-${ipv6_prefix}::fa"
    echo "IPv6 address range: ${address_range_v6}"
    [ -n "${address_ranges}" ] && address_ranges+="\n"
    address_ranges+="- ${address_range_v6}"
fi

if [ -z "${address_ranges}" ]; then
    echo "Error: No valid IP ranges found for IP_FAMILY=${IP_FAMILY}"
    exit 1
fi

# Apply MetalLB IPAddressPool and L2Advertisement
kubectl apply -f - <<EOF
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  namespace: metallb-system
  name: kube-services
spec:
  addresses:
$(echo -e "${address_ranges}" | sed 's/^/    /')
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
