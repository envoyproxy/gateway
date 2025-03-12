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

   {{< tabpane text=true >}}
   {{% tab header="With External LoadBalancer Support" %}}

   You can also test the same functionality by sending traffic to the External IP. To get the external IP of the
   Envoy service, run:

   ```shell
   export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
   ```

   **Note**: In certain environments, the load balancer may be exposed using a hostname, instead of an IP address. If so, replace
   `ip` in the above command with `hostname`.

   Curl the example app through Envoy proxy:

   ```shell
   curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST/get
   ```

   {{% /tab %}}
   {{% tab header="Without LoadBalancer Support" %}}

   Get the name of the Envoy service created by the example Gateway:

   ```shell
   export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
   ```

   Port forward to the Envoy service:

   ```shell
   kubectl -n envoy-gateway-system port-forward service/${ENVOY_SERVICE} 8888:80 &
   ```

   Curl the example app through Envoy proxy:

   ```shell
   curl --verbose --header "Host: www.example.com" http://localhost:8888/get
   ```

   {{% /tab %}}
   {{< /tabpane >}}

</details>

---
