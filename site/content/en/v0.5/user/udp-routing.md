---
title: "UDP Routing"
---

The [UDPRoute][] resource allows users to configure UDP routing by matching UDP traffic and forwarding it to Kubernetes
backends. This guide will use CoreDNS example to walk you through the steps required to configure UDPRoute on Envoy
Gateway.

__Note:__ UDPRoute allows Envoy Gateway to operate as a non-transparent proxy between a UDP client and server. The lack
of transparency means that the upstream server will see the source IP and port of the Gateway instead of the client.
For additional information, refer to Envoy's [UDP proxy documentation][].

## Prerequisites

Install Envoy Gateway:

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.5.0 -n envoy-gateway-system --create-namespace
```

Wait for Envoy Gateway to become available:

```shell
kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
```

## Installation

Install CoreDNS in the Kubernetes cluster as the example backend. The installed CoreDNS is listening on
 UDP port 53 for DNS lookups.

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/v0.5.0/examples/kubernetes/udp-routing-example-backend.yaml
```

Wait for the CoreDNS deployment to become available:

```shell
kubectl wait --timeout=5m deployment/coredns --for=condition=Available
```

Update the Gateway from the Quickstart guide to include a UDP listener that listens on UDP port `5300`:

```shell
kubectl patch gateway eg --type=json --patch '[{
   "op": "add",
   "path": "/spec/listeners/-",
   "value": {
      "name": "coredns",
      "protocol": "UDP",
      "port": 5300,
      "allowedRoutes": {
         "kinds": [{
            "kind": "UDPRoute"
          }]
      }
    },
}]'
```

Verify the Gateway status:

```shell
kubectl get gateway/eg -o yaml
```

## Configuration

Create a UDPRoute resource to route UDP traffic received on Gateway port 5300 to the CoredDNS backend.

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: UDPRoute
metadata:
  name: coredns
spec:
  parentRefs:
    - name: eg
      sectionName: coredns
  rules:
    - backendRefs:
        - name: coredns
          port: 53
EOF
```

Verify the UDPRoute status:

```shell
kubectl get udproute/coredns -o yaml
```

## Testing

Get the External IP of the Gateway:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Use `dig` command to query the dns entry foo.bar.com through the Gateway.

```shell
dig @${GATEWAY_HOST} -p 5300 foo.bar.com
```

You should see the result of the dns query as the below output, which means that the dns query has been successfully
routed to the backend CoreDNS.

Note: 49.51.177.138 is the resolved address of GATEWAY_HOST.

```bash
; <<>> DiG 9.18.1-1ubuntu1.1-Ubuntu <<>> @49.51.177.138 -p 5300 foo.bar.com
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 58125
;; flags: qr aa rd; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 3
;; WARNING: recursion requested but not available

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 1232
; COOKIE: 24fb86eba96ebf62 (echoed)
;; QUESTION SECTION:
;foo.bar.com.			IN	A

;; ADDITIONAL SECTION:
foo.bar.com.		0	IN	A	10.244.0.19
_udp.foo.bar.com.	0	IN	SRV	0 0 42376 .

;; Query time: 1 msec
;; SERVER: 49.51.177.138#5300(49.51.177.138) (UDP)
;; WHEN: Fri Jan 13 10:20:34 UTC 2023
;; MSG SIZE  rcvd: 114
```

## Clean-Up

Follow the steps from the [Quickstart Guide](../quickstart) to uninstall Envoy Gateway.

Delete the CoreDNS example manifest and the UDPRoute:

```shell
kubectl delete deploy/coredns
kubectl delete service/coredns
kubectl delete cm/coredns
kubectl delete udproute/coredns
```

## Next Steps

Checkout the [Developer Guide](../../contributions/develop/) to get involved in the project.

[UDPRoute]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1alpha2.UDPRoute
[UDP proxy documentation]: https://www.envoyproxy.io/docs/envoy/v0.5.0/configuration/listeners/udp_filters/udp_proxy
