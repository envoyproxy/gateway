#!/usr/bin/env bash

set -euo pipefail

# Setup default values
CLUSTER_NAME=${CLUSTER_NAME:-"envoy-gateway"}
METALLB_VERSION=${METALLB_VERSION:-"v0.13.10"}
KIND_NODE_TAG=${KIND_NODE_TAG:-"v1.33.0"}
NUM_WORKERS=${NUM_WORKERS:-""}
IP_FAMILY=${IP_FAMILY:-"ipv4"}
CUSTOM_CNI=${CUSTOM_CNI:-"false"}

if [ "$CUSTOM_CNI" = "true" ]; then
  CNI_CONFIG="disableDefaultCNI: true"
else
  CNI_CONFIG="disableDefaultCNI: false"
fi

KIND_CFG=$(cat <<-EOM
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  ${CNI_CONFIG}
  ipFamily: ${IP_FAMILY}
  # uncomment following line when use IPv6 on macos or windows
  # apiServerAddress: 127.0.0.1
  # it's to prevent inherit search domains from the host which slows down DNS resolution
  # and cause problems to IPv6 only clusters running on IPv4 host.
  dnsSearch: []
nodes:
- role: control-plane
  labels:
    "topology.kubernetes.io/zone": 0
EOM
)

# https://kind.sigs.k8s.io/docs/user/quick-start/#multi-node-clusters
if [[ -n "${NUM_WORKERS}" ]]; then
for i in $(seq 1 "${NUM_WORKERS}"); do
  KIND_CFG+=$(printf "\n- role: worker\n  labels:\n    \"topology.kubernetes.io/zone\": %s" "$i")
done
fi

## Check if kind cluster already exists.
if go tool kind get clusters | grep -q "${CLUSTER_NAME}"; then
  echo "Cluster ${CLUSTER_NAME} already exists."
else
## Create kind cluster.
if [[ -z "${KIND_NODE_TAG}" ]]; then
  cat << EOF | go tool kind create cluster --name "${CLUSTER_NAME}" --config -
${KIND_CFG}
EOF
else
  cat << EOF | go tool kind create cluster --image "kindest/node:${KIND_NODE_TAG}" --name "${CLUSTER_NAME}" --config -
${KIND_CFG}
EOF
fi
fi
if [ "$CUSTOM_CNI" = "true" ]; then
## Install Calico
# Determine the operating system
OS=$(uname -s)
case $OS in
    Darwin)
        CILIUM_CLI_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/cilium-cli/main/stable.txt)
        CLI_ARCH=amd64
        if [ "$(uname -m)" = "arm64" ]; then CLI_ARCH=arm64; fi
        curl -L --fail --remote-name-all "https://github.com/cilium/cilium-cli/releases/download/${CILIUM_CLI_VERSION}/cilium-darwin-${CLI_ARCH}.tar.gz"{,.sha256sum}
        shasum -a 256 -c cilium-darwin-${CLI_ARCH}.tar.gz.sha256sum
        tar xf cilium-darwin-${CLI_ARCH}.tar.gz
        rm cilium-darwin-${CLI_ARCH}.tar.gz{,.sha256sum}
        ;;
    Linux)
        CILIUM_CLI_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/cilium-cli/main/stable.txt)
        CLI_ARCH=amd64
        if [ "$(uname -m)" = "aarch64" ]; then CLI_ARCH=arm64; fi
        curl -L --fail --remote-name-all "https://github.com/cilium/cilium-cli/releases/download/${CILIUM_CLI_VERSION}/cilium-linux-${CLI_ARCH}.tar.gz"{,.sha256sum}
        sha256sum --check cilium-linux-${CLI_ARCH}.tar.gz.sha256sum
        tar xf cilium-linux-${CLI_ARCH}.tar.gz
        rm cilium-linux-${CLI_ARCH}.tar.gz{,.sha256sum}
        ;;
    *)
        echo "Unsupported operating system: $OS"
        exit 1
        ;;
esac
mkdir -p bin
chmod +x cilium
mv cilium bin
fi

## Install MetalLB.
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/"${METALLB_VERSION}"/config/manifests/metallb-native.yaml
needCreate="$(kubectl get secret -n metallb-system memberlist --no-headers --ignore-not-found -o custom-columns=NAME:.metadata.name)"
if [ -z "$needCreate" ]; then
    kubectl create secret generic -n metallb-system memberlist --from-literal=secretkey="$(openssl rand -base64 128)"
fi


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

apply_metallb_ranges() {
kubectl apply -f - <<EOF >/dev/null 2>&1
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
}

RETRY_INTERVAL=5  # seconds
TIMEOUT=120        # seconds
ELAPSED_TIME=0

if [ "$CUSTOM_CNI" = "true" ]; then
  CILIUM_BIN="./bin/cilium"
  $CILIUM_BIN install --wait --version 1.16.4
  $CILIUM_BIN status --wait
fi

# Apply MetalLB IPAddressPool and L2Advertisement
echo "Applying configuration with retries..."
  # Retry loop
  while [ $ELAPSED_TIME -lt $TIMEOUT ]; do
    if apply_metallb_ranges; then
      echo "Configuration applied successfully."
      exit 0
    else
      echo "Trying to apply configuration. Retrying in $RETRY_INTERVAL seconds..."
    fi
    sleep $RETRY_INTERVAL
    ELAPSED_TIME=$((ELAPSED_TIME + RETRY_INTERVAL))
  done
