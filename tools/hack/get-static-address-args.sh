#!/usr/bin/env bash

# This script outputs --usable-address and --unusable-address flags for the
# GatewayStaticAddresses conformance test. It computes a usable IP address
# from the Docker kind network (within the MetalLB pool) and uses a hostname
# as the unusable address since hostname addresses cannot be assigned via
# ExternalIPs or LoadBalancer.
#
# Usage: get-static-address-args.sh <ip-family>
# Output: --usable-address=<ip> --unusable-address=<hostname>

set -euo pipefail

IP_FAMILY="${1:-ipv4}"
UNUSABLE_ADDRESS="unusable.example.com"

if [ "${IP_FAMILY}" = "ipv6" ]; then
    subnet_v6=$(docker network inspect kind 2>/dev/null | jq -r '.[].IPAM.Config[] | select(.Subnet | contains(":")) | .Subnet' | head -1)
    if [ -z "${subnet_v6}" ]; then
        exit 0
    fi
    ipv6_prefix="${subnet_v6%::*}"
    # Use an address within the MetalLB pool range (c8-fa = .200-.250)
    echo "--usable-address=${ipv6_prefix}::fa --unusable-address=${UNUSABLE_ADDRESS}"
else
    subnet_v4=$(docker network inspect kind 2>/dev/null | jq -r '.[].IPAM.Config[] | select(.Subnet | contains(":") | not) | .Subnet' | head -1)
    if [ -z "${subnet_v4}" ]; then
        exit 0
    fi
    address_prefix_v4=$(echo "${subnet_v4}" | awk -F. '{print $1"."$2"."$3}')
    # Use an address within the MetalLB pool range (.200-.250)
    echo "--usable-address=${address_prefix_v4}.250 --unusable-address=${UNUSABLE_ADDRESS}"
fi
