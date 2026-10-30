---
title: "Backend Mutual TLS with SPIRE"
---

This task shows how Envoy Gateway can establish mTLS to a backend using [SPIFFE][] identities issued by [SPIRE][].
Instead of storing certificates in Kubernetes Secrets, Envoy fetches its client certificate and the trust bundle from the
SPIRE Agent at runtime through the [Secret Discovery Service (SDS)][SDS]. SPIRE rotates the certificates automatically, and
Envoy picks up the new ones without a restart.

The backend in this task is part of an Istio service mesh that also uses SPIRE as its identity provider. Because the
Gateway and the mesh workloads share the same trust domain, Envoy Gateway can talk to mesh workloads with mTLS
without joining the mesh.

## Prerequisites

- A Kubernetes cluster.
- [Helm][] and [istioctl][].
- Envoy Gateway installed. Follow the [Quickstart](../../quickstart) to install it.

## Install SPIRE

Install SPIRE with the trust domain `example.org`:

```shell
helm upgrade --install -n spire-server spire-crds spire-crds \
  --repo https://spiffe.github.io/helm-charts-hardened/ --create-namespace
helm upgrade --install -n spire-server spire spire \
  --repo https://spiffe.github.io/helm-charts-hardened/ --wait --timeout 10m \
  --set global.spire.trustDomain="example.org"
```

The `spiffe-oidc-discovery-provider` pod may restart a few times while the SPIRE Agents start up. This is expected,
and the longer `--timeout` gives it time to recover.

The chart installs:

- The SPIRE Agent as a DaemonSet. It serves the Workload API and SDS on a Unix domain socket on every node.
- The [SPIFFE CSI driver][] (`csi.spiffe.io`), which mounts that socket into pods.
- A default `ClusterSPIFFEID` that gives every pod outside the SPIRE namespaces the identity
  `spiffe://example.org/ns/<namespace>/sa/<service-account>`. It has no pod selector, so pods don't need any label.

The Envoy proxy pods get their identity from this default registration, so you don't need to register them separately.
If you installed SPIRE with a custom `ClusterSPIFFEID` that selects pods by label, add that label to the Envoy pods with
`spec.provider.kubernetes.envoyDeployment.pod.labels` in the [EnvoyProxy][] resource.

To confirm which pods the default registration selects, run:

```shell
kubectl get clusterspiffeid spire-server-spire-default -o jsonpath='{.spec}'
```

## Install Istio and the backend

Install Istio with a `spire` sidecar injection template, so mesh workloads get their certificates from SPIRE instead of
istiod:

```shell
cat <<EOF | istioctl install -y -f -
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  namespace: istio-system
spec:
  profile: default
  meshConfig:
    trustDomain: example.org
  values:
    sidecarInjectorWebhook:
      templates:
        spire: |
          spec:
            initContainers:
            - name: istio-proxy
              volumeMounts:
              - name: workload-socket
                mountPath: /run/secrets/workload-spiffe-uds
                readOnly: true
            volumes:
              - name: workload-socket
                csi:
                  driver: "csi.spiffe.io"
                  readOnly: true
EOF
```

Create a namespace with sidecar injection enabled, then deploy `httpbin` into it:

```shell
kubectl create namespace spire-sample
kubectl label namespace spire-sample istio-injection=enabled

cat <<EOF | kubectl apply -n spire-sample -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: httpbin
---
apiVersion: v1
kind: Service
metadata:
  name: httpbin
  labels:
    app: httpbin
spec:
  ports:
  - name: http
    port: 8000
    targetPort: 8080
  selector:
    app: httpbin
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: httpbin
spec:
  replicas: 1
  selector:
    matchLabels:
      app: httpbin
  template:
    metadata:
      labels:
        app: httpbin
      annotations:
        inject.istio.io/templates: "sidecar,spire"
    spec:
      serviceAccountName: httpbin
      containers:
      - name: httpbin
        image: docker.io/mccutchen/go-httpbin:v2.15.0
        ports:
        - containerPort: 8080
EOF
```

The `httpbin` workload's SPIFFE ID is `spiffe://example.org/ns/spire-sample/sa/httpbin`.

## Enable SDS Secret references

Envoy Gateway reads SDS settings from Secrets of type `gateway.envoyproxy.io/sds`. This feature is turned off by
default. To turn it on, set `enableSDSSecretRef` under `extensionApis` in the [EnvoyGateway][] configuration.

If you installed Envoy Gateway with Helm, as in the [Quickstart](../../quickstart), use `helm upgrade` with
`--reuse-values` so your other settings stay as they are:

```shell
helm upgrade eg oci://docker.io/envoyproxy/gateway-helm --version {{< helm-version >}} \
  -n envoy-gateway-system --reuse-values \
  --set config.envoyGateway.extensionApis.enableSDSSecretRef=true
```

Otherwise, edit the `envoy-gateway-config` ConfigMap with `kubectl edit configmap envoy-gateway-config -n envoy-gateway-system`
and add the setting to `envoy-gateway.yaml`. Keep all of the existing configuration:

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
    ... keep the existing configuration ...
    extensionApis:
      ... keep any existing extensionApis settings ...
      enableSDSSecretRef: true
```

{{< boilerplate rollout-envoy-gateway >}}

```shell
kubectl -n envoy-gateway-system rollout restart deployment envoy-gateway
kubectl -n envoy-gateway-system rollout status deployment envoy-gateway
```

## Create the SDS reference Secrets

An SDS reference Secret holds two keys:

- `url`: the address of the SDS server. Here it is the SPIRE Agent socket that is mounted into the Envoy pod.
- `secretName`: the name of the SDS resource to request. The SPIRE Agent serves two resources:
  - `default`: the workload's X.509 SVID (certificate and private key).
  - `ROOTCA`: the trust bundle for the trust domain.

Create one Secret for Envoy's client certificate, and one for the CA that validates the backend:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: spire-svid
  namespace: default
type: gateway.envoyproxy.io/sds
stringData:
  url: unix:///run/secrets/workload-spiffe-uds/socket
  secretName: default
---
apiVersion: v1
kind: Secret
metadata:
  name: spire-bundle
  namespace: spire-sample
type: gateway.envoyproxy.io/sds
stringData:
  url: unix:///run/secrets/workload-spiffe-uds/socket
  secretName: ROOTCA
EOF
```

These Secrets hold no key material. The certificates never leave the node: Envoy fetches them directly from the
SPIRE Agent.

## Configure the Gateway

Create an [EnvoyProxy][] resource that:

- mounts the SPIRE Agent socket into the Envoy container with the SPIFFE CSI driver,
- adds an init container that waits until the socket is available, so Envoy doesn't start before it can reach SDS,
- uses the `spire-svid` Secret as the client certificate for all backend TLS connections.

Then create a Gateway that references it:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: spire-proxy-config
  namespace: default
spec:
  backendTLS:
    clientCertificateRef:
      kind: Secret
      name: spire-svid
    alpnProtocols:
      - istio
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        pod:
          volumes:
            - name: workload-socket
              csi:
                driver: "csi.spiffe.io"
                readOnly: true
        container:
          volumeMounts:
            - name: workload-socket
              mountPath: /run/secrets/workload-spiffe-uds
              readOnly: true
        patch:
          type: StrategicMerge
          value:
            spec:
              template:
                spec:
                  initContainers:
                    - name: wait-for-spire-socket
                      image: docker.io/library/busybox:1.36
                      volumeMounts:
                        - name: workload-socket
                          mountPath: /run/secrets/workload-spiffe-uds
                          readOnly: true
                      command:
                        - sh
                        - -c
                        - |
                          until [ -S /run/secrets/workload-spiffe-uds/socket ]; do
                            echo "Waiting for SPIRE Agent socket..."
                            sleep 1
                          done
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: eg
  namespace: default
spec:
  gatewayClassName: eg
  infrastructure:
    parametersRef:
      group: gateway.envoyproxy.io
      kind: EnvoyProxy
      name: spire-proxy-config
  listeners:
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: All
EOF
```

`alpnProtocols: [istio]` tells the Istio sidecar on the backend that the incoming connection is Istio mTLS. If you
leave it out, the sidecar only accepts the connection when mTLS is enforced with a `STRICT` [PeerAuthentication][].

## Route traffic to the backend with mTLS

Create an HTTPRoute to `httpbin`, and a [BackendTLSPolicy][] that validates the backend's certificate against the
SPIRE trust bundle and its SPIFFE ID:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: httpbin
  namespace: spire-sample
spec:
  parentRefs:
    - name: eg
      namespace: default
  rules:
    - backendRefs:
        - name: httpbin
          port: 8000
---
apiVersion: gateway.networking.k8s.io/v1
kind: BackendTLSPolicy
metadata:
  name: httpbin
  namespace: spire-sample
spec:
  targetRefs:
    - group: ""
      kind: Service
      name: httpbin
  validation:
    caCertificateRefs:
      - group: ""
        kind: Secret
        name: spire-bundle
    hostname: httpbin.spire-sample.svc.cluster.local
    subjectAltNames:
      - type: URI
        uri: spiffe://example.org/ns/spire-sample/sa/httpbin
EOF
```

Istio also runs Gateway API controllers, so the BackendTLSPolicy status shows extra ancestors from
`istio.io/gateway-controller` and `istio.io/mesh-controller` with `Accepted=False`, because Istio doesn't support
Secret CA references. You can ignore these. The ancestor from `gateway.envoyproxy.io/gatewayclass-controller` should
be `Accepted=True`:

```shell
kubectl -n spire-sample get backendtlspolicy httpbin \
  -o jsonpath='{range .status.ancestors[*]}{.controllerName}: {.conditions[0].type}={.conditions[0].status}{"\n"}{end}'
```

## Testing

Get the Gateway address:

```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

If your cluster has no LoadBalancer support, port-forward to the Envoy service instead:

```shell
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
kubectl -n envoy-gateway-system port-forward service/${ENVOY_SERVICE} 8888:80 &
export GATEWAY_HOST=localhost:8888
```

Send a request through the Gateway:

```shell
curl -s http://$GATEWAY_HOST/headers
```

The Istio sidecar on `httpbin` adds an `X-Forwarded-Client-Cert` header that describes the client certificate it
received. The `URI` field shows the SPIFFE ID that SPIRE issued to the Envoy proxy:

```json
{
  "headers": {
    "X-Forwarded-Client-Cert": [
      "By=spiffe://example.org/ns/spire-sample/sa/httpbin;Hash=...;Subject=\"O=SPIRE,C=US\";URI=spiffe://example.org/ns/envoy-gateway-system/sa/envoy-default-eg-e41e7b31"
    ],
    ...
  }
}
```

This confirms that both sides authenticated each other with SPIRE-issued certificates.

## Clean up

```shell
kubectl delete httproute/httpbin backendtlspolicy/httpbin secret/spire-bundle -n spire-sample
kubectl delete gateway/eg envoyproxy/spire-proxy-config secret/spire-svid -n default
kubectl delete namespace spire-sample
istioctl uninstall --purge -y
helm uninstall -n spire-server spire spire-crds
kubectl delete crd clusterfederatedtrustdomains.spire.spiffe.io clusterspiffeids.spire.spiffe.io clusterstaticentries.spire.spiffe.io
kubectl delete namespace spire-server
```

The SPIRE CRDs are kept when you uninstall the `spire-crds` chart, so delete them explicitly.

If you keep Envoy Gateway installed, turn SDS Secret references off again, so that Envoy Gateway stops accepting
`gateway.envoyproxy.io/sds` Secrets. With Helm:

```shell
helm upgrade eg oci://docker.io/envoyproxy/gateway-helm --version {{< helm-version >}} \
  -n envoy-gateway-system --reuse-values \
  --set config.envoyGateway.extensionApis.enableSDSSecretRef=false
kubectl -n envoy-gateway-system rollout restart deployment envoy-gateway
```

Otherwise, remove `enableSDSSecretRef` from the `envoy-gateway-config` ConfigMap (or set it back to `false`), leave
the other settings unchanged, and restart Envoy Gateway:

```shell
kubectl edit configmap envoy-gateway-config -n envoy-gateway-system
kubectl -n envoy-gateway-system rollout restart deployment envoy-gateway
```

## Security considerations

Anyone who can create or reference a `gateway.envoyproxy.io/sds` Secret can make Envoy request any SDS resource that
the proxy's SPIFFE identity is allowed to fetch. Turn on `enableSDSSecretRef` only when the users who can create
these Secrets are trusted. To keep mutually untrusted tenants apart, give them separate Envoy proxy deployments, and
so separate SPIFFE identities. See [SDS Certificate References][] for more on SDS Secret references.

[SPIFFE]: https://spiffe.io/
[SPIRE]: https://spiffe.io/docs/latest/spire-about/
[SDS]: https://www.envoyproxy.io/docs/envoy/latest/configuration/security/secret
[SPIFFE CSI driver]: https://github.com/spiffe/spiffe-csi
[EnvoyGateway]: ../../api/extension_types#envoygateway
[Helm]: https://helm.sh/docs/intro/install/
[istioctl]: https://istio.io/latest/docs/setup/getting-started/#download
[EnvoyProxy]: ../../api/extension_types#envoyproxy
[SDS Certificate References]: ../tls-termination/#sds-certificate-references
[BackendTLSPolicy]: https://gateway-api.sigs.k8s.io/reference/api-types/policy/backendtlspolicy/
[PeerAuthentication]: https://istio.io/latest/docs/reference/config/security/peer_authentication/
