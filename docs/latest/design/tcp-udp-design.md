# TCP&UDP Proxy Design 

Even though most of the use cases for Envoy Gateway are at Layer-7, Envoy Gateway can also work at Layer-4 to proxy TCP 
and UDP traffic. This document will explore the options we have when operating Envoy Gateway at Layer-4 and explain the 
design decision.

## Non-transparent Proxy Mode
Envoy terminates the downstream connection, connects the upstream with its own address,  and proxies the TCP/UDP traffic
from the downstream to the upstream. In this mode, the upstream will see Envoy's address.

## Transparent Proxy Mode
Envoy terminates the downstream connection, connects the upstream with the downstream source address,  and proxies the
TCP/UDP traffic from the downstream to the upstream. In this mode, the upstream will see the downstream original address.

## The Implication of Transparent Proxy Mode

### Escalated Privilege
Envoy needs to bind to the downstream source IP when connecting to the upstream, which means Envoy requires 
escalated CAP_NET_ADMIN privileges. This is often considered as a bad security practice and not allowed in a sensitive 
deployment.

### Routing
The upstream can see the original source IP, but the original port number can't pass through, so the return traffic from
the upstream must be routed back to Envoy because only Envoy knows how to send the return traffic back to the right port
number of the downstream, which requires routing at the upstream side be deliberately set up. In a Kubernetes cluster, 
Envoy Gateway will have to carefully cooperate with a CNI plugin to get the routing right.

## The Design Decision(For Now)

Envoy Gateway will only support proxying in non-transparent mode for now. As we expand our use cases to non-Kubernetes 
environments, we'll look back and try to implement the transparent mode.

