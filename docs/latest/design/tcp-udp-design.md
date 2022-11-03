# TCP&UDP Proxy Design 

Even though most of the use cases for Envoy Gateway are at Layer-7, Envoy Gateway can also work at Layer-4 to proxy TCP 
and UDP traffic. This document will explore the options we have when operating Envoy Gateway at Layer-4 and explain the 
design decision.

Envoy can work as a non-transparent proxy or a [transparent proxy](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/other_features/ip_transparency#arch-overview-ip-transparency-original-src-listener), 
so ideally, Envoy Gateway should also be able to work in these two modes:

## Non-transparent Proxy Mode
Envoy terminates the downstream connection, connects the upstream with its own address,  and proxies the TCP/UDP traffic
from the downstream to the upstream. In this mode, the upstream will see Envoy's address.

## Transparent Proxy Mode
Envoy terminates the downstream connection, connects the upstream with the downstream source address,  and proxies the
TCP/UDP traffic from the downstream to the upstream. In this mode, the upstream will see the original downstream 
address.

## The Implications of Transparent Proxy Mode

### Escalated Privilege
Envoy needs to bind to the downstream source IP when connecting to the upstream, which means Envoy requires 
escalated CAP_NET_ADMIN privileges. This is often considered as a bad security practice and not allowed in some 
sensitive deployments.

### Routing
The upstream can see the original source IP, but the original port number won't be past through, so the return 
traffic from the upstream must be routed back to Envoy because only Envoy knows how to send the return traffic back 
to the right port number of the downstream, which requires routing at the upstream side must be deliberately set up. 
In a Kubernetes cluster, Envoy Gateway will have to carefully cooperate with CNI plugins to get the routing right.

## The Design Decision(For Now)

Since now Envoy Gateway's primary goal is implementing an in-cluster Gateway, and it's tricky to support non-transparent 
mode in cluster networks, Envoy Gateway currently will only support proxying in non-transparent mode. As we expand our 
use cases to non-Kubernetes environments, we'll look back and see what we can do to implement the transparent mode.