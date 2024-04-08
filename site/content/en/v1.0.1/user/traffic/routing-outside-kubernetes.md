---
title: "Routing outside Kubernetes"
---

Routing to endpoints outside the Kubernetes cluster where Envoy Gateway and its corresponding Envoy Proxy fleet is
running is a common use. This can be achieved by defining FQDN addresses in a [EndpointSlice][].

## Installation

Follow the steps from the [Quickstart](../../quickstart) guide to install Envoy Gateway and the example manifest.
Before proceeding, you should be able to query the example backend using HTTP.

## Configuration

* Define a Service and EndpointSlice that represents https://httpbin.org

    Apply the following resources to your cluster:

    ```yaml
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: httpbin
      namespace: default
    spec:
      ports:
        - port: 443
          protocol: TCP
          targetPort: 443
          name: https
    ---
    apiVersion: discovery.k8s.io/v1
    kind: EndpointSlice
    metadata:
      name: httpbin
      namespace: default
      labels:
        kubernetes.io/service-name: httpbin 
    addressType: FQDN
    ports:
    - name: https
      protocol: TCP
      port: 443
    endpoints:
    - addresses:
      - "httpbin.org"
    ```

* Update the [Gateway][] to include a TLS Listener on port 443

    ```shell
    kubectl patch gateway eg --type=json --patch '[{
      "op": "add",
      "path": "/spec/listeners/-",
      "value": {
          "name": "tls",
          "protocol": "TLS",
          "port": 443,
          "tls": {
            "mode": "Passthrough",
          },
        },
    }]'
    ```

* Add a [TLSRoute][] that can route incoming traffic to the above backend that we created

    Apply the following resource to your cluster:

    ```yaml
    ---
    apiVersion: gateway.networking.k8s.io/v1alpha2
    kind: TLSRoute
    metadata:
      name: httpbin 
    spec:
      parentRefs:
      - name: eg 
        sectionName: tls
      rules:
      - backendRefs:
        - name: httpbin
          port: 443
    ```    

Obtain the Gateway address:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Send a request and view the response:

```shell
curl -I -HHost:httpbin.org --resolve "httpbin.org:443:${GATEWAY_HOST}" https://httpbin.org:443
```

[EndpointSlice]: https://kubernetes.io/docs/concepts/services-networking/endpoint-slices/
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway/
[TLSRoute]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TLSRoute
