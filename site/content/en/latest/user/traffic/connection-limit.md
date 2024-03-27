---
title: "Connection Limit"
---

The connection limit features allows users to limit the number of concurrently active TCP connections on a [Gateway][] or a [Listener][].
When the [connection limit][] is reached, new connections are closed immediately by Envoy proxy. It's possible to configure a delay for connection rejection.

Users may want to limit the number of connections for several reasons:
* Protect resources like CPU and Memory.
* Ensure that different listeners can receive a fair share of global resources.
* Protect from malicious activity like DoS attacks. 

Envoy Gateway introduces a new CRD called [Client Traffic Policy][] that allows the user to describe their desired connection limit settings.
This instantiated resource can be linked to a [Gateway][].

The Envoy [connection limit][] implementation is distributed: counters are not synchronized between different envoy proxies.

When a [Client Traffic Policy][] is attached to a gateway, the connection limit will apply differently based on the 
[Listener][] protocol in use: 
- HTTP: all HTTP listeners in a [Gateway][] will share a common connection counter, and a limit defined by the policy.
- HTTPS/TLS: each HTTPS/TLS listener will have a dedicated connection counter, and a limit defined by the policy.


## Prerequisites

### Install Envoy Gateway

* Follow the steps from the [Quickstart Guide](../../quickstart) to install Envoy Gateway and the HTTPRoute example manifest.
  Before proceeding, you should be able to query the example backend using HTTP.

### Install the hey load testing tool
* The `hey` CLI will be used to generate load and measure response times. Follow the installation instruction from the [Hey project] docs.

## Test and customize connection limit settings

This example we use `hey` to open 10 connections and execute 1 RPS per connection for 10 seconds.

```shell
hey -c 10 -q 1 -z 10s  -host "www.example.com" http://${GATEWAY_HOST}/get
```

```console
Summary:
  Total:	10.0058 secs
  Slowest:	0.0275 secs
  Fastest:	0.0029 secs
  Average:	0.0111 secs
  Requests/sec:	9.9942

[...]

Status code distribution:
  [200]	100 responses
```

There are no connection limits, and so all 100 requests succeed. 

Next, we apply a limit of 5 connections. 

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: connection-limit-ctp
  namespace: default
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
    namespace: default
  connection:
    connectionLimit:
      value: 5    
EOF
```

Execute the load simulation again.

```shell
hey -c 10 -q 1 -z 10s  -host "www.example.com" http://${GATEWAY_HOST}/get
```

```console
Summary:
  Total:	11.0327 secs
  Slowest:	0.0361 secs
  Fastest:	0.0013 secs
  Average:	0.0088 secs
  Requests/sec:	9.0640

[...] 

Status code distribution:
  [200]	50 responses

Error distribution:
  [50]	Get "http://localhost:8888/get": EOF
```

With the new connection limit, only 5 of 10 connections are established, and so only 50 requests succeed.  


[Client Traffic Policy]: ../../../api/extension_types#clienttrafficpolicy
[Hey project]: https://github.com/rakyll/hey
[connection limit]: https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/network_filters/connection_limit_filter
[listener]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Listener
[gateway]:  https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.Gateway
