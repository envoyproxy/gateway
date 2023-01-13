# UDP Routing

The [UDPRoute][] resource allows users to configure UDP routing by matching UDP traffic and forwarding it to
Kubernetes backends. To learn more about UDP routing, refer to the [Gateway API documentation][].

Follow the steps from the [Quickstart Guide](quickstart.md) to install Envoy Gateway and then install the example
resources used for this guide.

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/udp-routing.yaml
```

The manifest installs a [GatewayClass][], [Gateway][], UDPRoute and a CoreDNS deployment and service for testing. The 
Gateway is listening on UDP port 5300 and forward the UDP traffic to the port 53 of the backend CoreDNS service.  

First, let's get the Gateway's address.
```shell
export GATEWAY_HOST=$(kubectl get gateway/udp-gateway -o jsonpath='{.status.addresses[0].value}')
```

Use `dig` command to query the dns entry foo.bar.com on port 5003 of the Gateway.

```shell
dig @${GATEWAY_HOST} -p 5300 foo.bar.com
```

You should see the result of the dns query as the below output, which means that the dns query has been successfully 
routed to the backend CoreDNS.

```bash
; <<>> DiG 9.18.1-1ubuntu1.1-Ubuntu <<>> @10.96.152.156 -p 5300 foo.bar.com
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
;; SERVER: 10.96.152.156#5300(10.96.152.156) (UDP)
;; WHEN: Fri Jan 13 10:20:34 UTC 2023
;; MSG SIZE  rcvd: 114
```

[UDPRoute]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1alpha2.UDPRoute/
[Gateway API documentation]: https://gateway-api.sigs.k8s.io/
[GatewayClass]: https://gateway-api.sigs.k8s.io/api-types/gatewayclass/
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway/
