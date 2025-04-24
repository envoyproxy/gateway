---
---

Follow the steps below to install Envoy Gateway and the example manifest. Before
proceeding, you should be able to query the example backend using HTTP.

<details>
<summary>Expand for instructions</summary>

1. Install the Gateway API CRDs and Envoy Gateway using Helm:

   ```shell
   helm install eg oci://docker.io/envoyproxy/gateway-helm --version {{< helm-version >}} -n envoy-gateway-system --create-namespace
   ```

2. Install the GatewayClass, Gateway, HTTPRoute and example app:

   ```shell
   kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/{{< yaml-version >}}/quickstart.yaml -n default
   ```

3. Verify Connectivity:
   {{< tabpane-include testing >}}
</details>
