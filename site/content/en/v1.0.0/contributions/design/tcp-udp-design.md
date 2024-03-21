---
title: "TCP and UDP Proxy Design "
---

Even though most of the use cases for Envoy Gateway are at Layer-7, Envoy Gateway can also work at Layer-4 to proxy TCP 
and UDP traffic. This document will explore the options we have when operating Envoy Gateway at Layer-4 and explain the 
design decision.

Envoy can work as a non-transparent proxy or a transparent proxy for both [TCP][]
 and [UDP][]
, so ideally, Envoy Gateway should also be able to work in these two modes:

## Non-transparent Proxy Mode
For TCP, Envoy terminates the downstream connection, connects the upstream with its own IP address, and proxies the 
TCP traffic from the downstream to the upstream. 

For UDP, Envoy receives UDP datagrams from the downstream, and uses its own IP address as the sender IP address when 
proxying the UDP datagrams to the upstream.

In this mode, the upstream will see Envoy's IP address and port.

## Transparent Proxy Mode
For TCP, Envoy terminates the downstream connection, connects the upstream with the downstream IP address, and proxies 
the TCP traffic from the downstream to the upstream. 

For UDP, Envoy receives UDP datagrams from the downstream, and uses the downstream IP address as the sender IP address 
when proxying the UDP datagrams to the upstream.

In this mode, the upstream will see the original downstream IP address and Envoy's mac address.

Note: Even in transparent mode, the upstream can't see the port number of the downstream because Envoy doesn't forward 
the port number.

## The Implications of Transparent Proxy Mode

### Escalated Privilege
Envoy needs to bind to the downstream IP when connecting to the upstream, which means Envoy requires escalated 
CAP_NET_ADMIN privileges. This is often considered as a bad security practice and not allowed in some sensitive deployments.

### Routing
The upstream can see the original source IP, but the original port number won't be passed, so the return 
traffic from the upstream must be routed back to Envoy because only Envoy knows how to send the return traffic back
to the right port number of the downstream, which requires routing at the upstream side to be set up. 
In a Kubernetes cluster, Envoy Gateway will have to carefully cooperate with CNI plugins to get the routing right.

## The Design Decision (For Now)

The implementation will only support proxying in non-transparent mode i.e. the backend will see the source IP and 
port of the deployed Envoy instance instead of the client.

[TCP]: https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/other_features/ip_transparency#arch-overview-ip-transparency-original-src-listener
[UDP]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/udp/udp_proxy/v3/udp_proxy.proto#envoy-v3-api-msg-extensions-filters-udp-udp-proxy-v3-udpproxyconfig
