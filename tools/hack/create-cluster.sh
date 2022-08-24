#!/bin/bash

set -euo pipefail

## Create kind cluster.
tools/bin/kind create cluster \
    --name envoy-gateway 


## Install metallb.
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/namespace.yaml
if ! kubectl get secret -n metallb-system memberlist; then
    kubectl create secret generic -n metallb-system memberlist --from-literal=secretkey="$(openssl rand -base64 128)"
fi
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/metallb.yaml
# Apply config with addresses based on docker network IPAM
subnet=$(docker network inspect kind | jq -r '.[].IPAM.Config[].Subnet | select(contains(":") | not)')
# Assume default kind network subnet prefix of 16, and choose addresses in that range.
address_first_octets=$(echo ${subnet} | awk -F. '{printf "%s.%s",$1,$2}')
address_range="${address_first_octets}.255.200-${address_first_octets}.255.250"
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    address-pools:
    - name: default
      protocol: layer2
      addresses:
      - ${address_range}
EOF
