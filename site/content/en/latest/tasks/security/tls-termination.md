---
title: "TLS Termination for TCP"
---

This task walks through configuring TLS Terminate mode for TCP traffic via Envoy Gateway. For HTTPS termination on an HTTP listener (TLS + HTTPRoute), see the [Secure Gateways][secure-gateways] task instead.

In both patterns below, Envoy terminates the client TLS connection on a `TLS` listener and forwards the decrypted bytes to a backend Service over plain TCP. The difference is how the listener selects a backend:

- **Single backend with [TCPRoute][]** — every connection accepted on the listener goes to one backend Service. TCPRoute does not inspect SNI, so each backend needs its own listener.
- **SNI-based routing with [TLSRoute][]** — one listener fans out to multiple backends, dispatched by the client's SNI. Many TLSRoutes can attach to the same listener, so you avoid the listener-per-backend explosion.

TLSRoute itself works with both listener TLS modes: this task covers `Terminate` (Envoy decrypts and forwards plain TCP); for `Passthrough` (Envoy SNI-routes the still-encrypted bytes, leaving termination to the backend) see the [TLS Passthrough][tls-passthrough] task.

This task uses a self-signed CA, so it should be used for testing and demonstration purposes only.

## Prerequisites

- OpenSSL to generate TLS assets.

## Installation

{{< boilerplate prerequisites >}}

## Single Backend with TCPRoute

This section configures a `TLS` listener that terminates TLS for `www.example.com` and forwards the decrypted bytes to a single backend Service via a [TCPRoute][]. Because TCPRoute does not match on SNI, this listener can only serve one backend — additional backends would each need their own listener.

### TLS Certificates

Generate the certificates and keys used by the Gateway to terminate client TLS connections.

Create a root certificate and private key to sign certificates:

```shell
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=example Inc./CN=example.com' -keyout example.com.key -out example.com.crt
```

Create a certificate and a private key for `www.example.com`:

```shell
openssl req -out www.example.com.csr -newkey rsa:2048 -nodes -keyout www.example.com.key -subj "/CN=www.example.com/O=example organization"
openssl x509 -req -days 365 -CA example.com.crt -CAkey example.com.key -set_serial 0 -in www.example.com.csr -out www.example.com.crt
```

Store the cert/key in a Secret:

```shell
kubectl create secret tls example-cert --key=www.example.com.key --cert=www.example.com.crt
```

Install the TLS Termination for TCP example resources:

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/tls-termination.yaml
```

Verify the Gateway status:

```shell
kubectl get gateway/eg -o yaml
```

### Testing

{{< tabpane text=true >}}
{{% tab header="With External LoadBalancer Support" %}}

Get the External IP of the Gateway:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Query the example app through the Gateway:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:443:${GATEWAY_HOST}" \
--cacert example.com.crt https://www.example.com/get
```

{{% /tab %}}
{{% tab header="Without LoadBalancer Support" %}}

Get the name of the Envoy service created the by the example Gateway:

```shell
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
```

Port forward to the Envoy service:

```shell
kubectl -n envoy-gateway-system port-forward service/${ENVOY_SERVICE} 8443:443 &
```

Query the example app through Envoy proxy:

```shell
curl -v -HHost:www.example.com --resolve "www.example.com:8443:127.0.0.1" \
--cacert example.com.crt https://www.example.com:8443/get
```

{{% /tab %}}
{{< /tabpane >}}

## SNI-Based Routing with TLSRoute

This section configures a single `TLS` listener that terminates TLS for `*.example.com` and dispatches each connection to a different backend Service based on the client's SNI, using [TLSRoute][]. Because the routes all attach to the same listener, this pattern is not bounded by the per-Gateway listener limit.

### TLS Certificates

Create a wildcard certificate that covers every hostname the routes will serve:

```shell
openssl req -out wildcard.example.com.csr -newkey rsa:2048 -nodes -keyout wildcard.example.com.key -subj "/CN=*.example.com/O=example organization"
openssl x509 -req -days 365 -CA example.com.crt -CAkey example.com.key -set_serial 1 -in wildcard.example.com.csr -out wildcard.example.com.crt
```

Store the cert/key in a Secret:

```shell
kubectl create secret tls wildcard-example-cert --key=wildcard.example.com.key --cert=wildcard.example.com.crt
```

### Additional Backend Service

The quickstart already deployed a `backend` Service. Add a second instance of the same echo application so the two TLSRoutes resolve to distinct pods and the SNI routing can be observed end-to-end:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: backend-2
---
apiVersion: v1
kind: Service
metadata:
  name: backend-2
  labels:
    app: backend-2
    service: backend-2
spec:
  ports:
    - name: http
      port: 3000
      targetPort: 3000
  selector:
    app: backend-2
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend-2
spec:
  replicas: 1
  selector:
    matchLabels:
      app: backend-2
      version: v1
  template:
    metadata:
      labels:
        app: backend-2
        version: v1
    spec:
      serviceAccountName: backend-2
      containers:
        - image: registry.k8s.io/gateway-api/echo-basic:v1.5.1
          imagePullPolicy: IfNotPresent
          name: backend-2
          ports:
            - containerPort: 3000
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
```

### Gateway and TLSRoutes

The Gateway exposes one `TLS` listener in `Terminate` mode. Two TLSRoutes attach to it via `sectionName: tls` and select traffic by hostname:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: eg
spec:
  gatewayClassName: eg
  listeners:
    - name: tls
      protocol: TLS
      port: 443
      hostname: "*.example.com"
      tls:
        mode: Terminate
        certificateRefs:
          - kind: Secret
            name: wildcard-example-cert
      allowedRoutes:
        namespaces:
          from: All
---
apiVersion: gateway.networking.k8s.io/v1
kind: TLSRoute
metadata:
  name: backend
spec:
  parentRefs:
    - name: eg
      sectionName: tls
  hostnames: ["backend.example.com"]
  rules:
    - backendRefs:
        - name: backend
          port: 3000
---
apiVersion: gateway.networking.k8s.io/v1
kind: TLSRoute
metadata:
  name: backend-2
spec:
  parentRefs:
    - name: eg
      sectionName: tls
  hostnames: ["backend-2.example.com"]
  rules:
    - backendRefs:
        - name: backend-2
          port: 3000
```

The listener certificate must cover every hostname the attached routes serve — either as a wildcard (as above) or via the cert's SAN list. Route hostnames must also match the listener's hostname pattern, otherwise the route will be rejected with a status condition.

### Testing

Get the Gateway address using either the LoadBalancer or port-forward variant shown in the [TCPRoute Testing](#testing) section above, then probe each backend by SNI:

```shell
curl -v -HHost:backend.example.com --resolve "backend.example.com:443:${GATEWAY_HOST}" \
--cacert example.com.crt https://backend.example.com/

curl -v -HHost:backend-2.example.com --resolve "backend-2.example.com:443:${GATEWAY_HOST}" \
--cacert example.com.crt https://backend-2.example.com/
```

The echo response includes the serving pod's name (`POD_NAME`), so each request lands on a different backend Service even though they share one listener.

## SDS Certificate References

Besides inline `kubernetes.io/tls` Secrets, a listener's `tls.certificateRefs` can reference a Secret of type `gateway.envoyproxy.io/sds`. Rather than embedding a certificate and key directly, this Secret tells Envoy to fetch the certificate at runtime from an external Secret Discovery Service (SDS) server, identified by two keys: `url` (the SDS server Unix domain socket path) and `secretName` (the resource name Envoy requests from that server):

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: sds-cert
type: gateway.envoyproxy.io/sds
stringData:
  url: /var/run/secrets/workload-spiffe-uds/socket
  secretName: default
```

A listener references it the same way it references any other certificate Secret:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: eg
spec:
  gatewayClassName: eg
  listeners:
    - name: tls
      protocol: TLS
      port: 443
      tls:
        mode: Terminate
        certificateRefs:
          - kind: Secret
            name: sds-cert
```

This feature is disabled by default. Add `enableSDSSecretRef` under `extensionApis` in the `EnvoyGateway` configuration stored in the `envoy-gateway-config` ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: envoy-gateway-config
  namespace: envoy-gateway-system
data:
  envoy-gateway.yaml: |
    apiVersion: gateway.envoyproxy.io/v1alpha1
    kind: EnvoyGateway
    gateway:
      controllerName: gateway.envoyproxy.io/gatewayclass-controller
    provider:
      type: Kubernetes
    extensionApis:
      enableSDSSecretRef: true
```

Preserve any other settings already present in `envoy-gateway.yaml`, then apply the ConfigMap and restart Envoy Gateway:

```shell
kubectl apply -f envoy-gateway-config.yaml
kubectl rollout restart deployment/envoy-gateway -n envoy-gateway-system
kubectl rollout status deployment/envoy-gateway -n envoy-gateway-system
```

The SDS Unix domain socket must also be mounted into the Envoy Proxy pod at the path specified by `url`. For example, the following `EnvoyProxy` mounts a socket directory provided on every Kubernetes node:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: sds-socket
  namespace: envoy-gateway-system
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        patch:
          type: StrategicMerge
          value:
            spec:
              template:
                spec:
                  volumes:
                    - name: sds-socket
                      hostPath:
                        path: /var/run/secrets/workload-spiffe-uds
                        type: Directory
                  containers:
                    - name: envoy
                      volumeMounts:
                        - name: sds-socket
                          mountPath: /var/run/secrets/workload-spiffe-uds
```

Attach the `EnvoyProxy` to the `GatewayClass` used by the Gateway:

```shell
kubectl apply -f envoyproxy-sds.yaml
kubectl patch gatewayclass eg --type=merge --patch '{"spec":{"parametersRef":{"group":"gateway.envoyproxy.io","kind":"EnvoyProxy","name":"sds-socket","namespace":"envoy-gateway-system"}}}'
```

Replace the `hostPath` with the directory used by the SDS provider. The directory must exist on every node that can run an Envoy Proxy pod. The Envoy process must have permission to connect to the socket. If the `GatewayClass` already references an `EnvoyProxy`, add the volume and mount to that resource instead of replacing `parametersRef`. Enabling SDS Secret references does not add the socket or change the proxy deployment automatically.

Kubernetes authorization and ReferenceGrant control access to the Secret containing the SDS connection details. The SDS server separately authorizes the `secretName` requested by Envoy using the proxy's SDS identity. Only enable SDS Secret references when users who can create or reference these Secrets are trusted to request the SDS resources available to that identity. Use separate Envoy Proxy deployments or SDS identities for mutually untrusted tenants.

Envoy waits for SDS-backed listener certificates before activating the listener. Envoy Gateway combines HTTPS listeners on the same address and port into one Envoy listener, so an unavailable SDS server or resource can delay activation or reset connections for every listener sharing that address and port.

To prevent HTTP/2 connection coalescing across HTTPS listeners that share a port, Envoy Gateway normally compares certificate DNS/SAN names and defaults listeners with overlapping certificates to HTTP/1.1 unless ALPN is explicitly configured. Because SDS-backed certificates are opaque to Envoy Gateway, every HTTPS listener sharing a port with an SDS-backed listener is treated conservatively in the same way, even when their configured hostnames are distinct.

[TCPRoute]: https://gateway-api.sigs.k8s.io/reference/api-spec/main/spec/#tcproute
[TLSRoute]: https://gateway-api.sigs.k8s.io/reference/api-spec/main/spec/#tlsroute
[tls-passthrough]: ../tls-passthrough/
[secure-gateways]: ../secure-gateways/
