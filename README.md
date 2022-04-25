# Envoy Gateway - KubeCon EU 2022 Demo

1. Build image

    ```shell
    REGISTRY=docker.io/envoyproxy PROJECT=envoy-gateway-controller VERSION=latest make container
    ```

2. Create kind cluster

    ```shell
    CLUSTERNAME=envoy-gateway make setup-kind-cluster
    ```

3. Load image into cluster

    ```shell
    kind --name envoy-gateway load docker-image docker.io/envoyproxy/envoy-gateway-controller:latest
    ```

4. Deploy envoy gateway provisioner

    ```shell
    kubectl apply -f examples/gateway-provisioner/
    ```

5. Create `GatewayClass`

    ```shell
    kubectl apply -f examples/envoy-gateway-demo/00-gatewayclass.yaml
    ```

    Verify it's been accepted
    ```shell
    kubectl describe gatewayclass envoy
    ```

    (Look for condition of `Accepted: true`.)

6. Create `Gateway`

    ```shell
    kubectl apply -f examples/envoy-gateway-demo/01-gateway.yaml
    ```

    Verify it's been provisioned
    ```shell
    kubectl get gateway envoy-gateway-1
    ```

    (Look for an address and a `Ready: true` condition, should come up in under a minute.)

7. Look at the infrastructure that's been provisioned

    ```shell
    kubectl get all
    ```

    (Look for two control plane pods and one envoy pod, all fully ready.)

8. Deploy echoserver workloads

    ```shell
    kubectl apply -f examples/envoy-gateway-demo/02-echoservers.yaml
    ```

9. Create `HTTPRoute`

    ```shell
    kubectl apply -f examples/envoy-gateway-demo/03-httproute.yaml
    ```

    Verify it's been accepted
    ```shell
    kubectl describe httproute echo-routes
    ```

    (Look for condition of `Accepted: true`.)

10. Make HTTP requests 

    **Notes:**
    - **the Gateway's address is assumed to be `172.18.255.200`, replace with actual address from the output of `kubectl get gateway envoy-gateway-1` if needed**
    - **this step assumes you are on a Linux host so the Gateway's address is routable. If on macOS, you'll need to `kubectl port-forward` to the Gateway**


    ```shell
    curl -i -H "Host: gateway.envoyproxy.io" 172.18.255.200/s1
    ```

    (Should route to echoserver-1)

    ```shell
    curl -i -H "Host: gateway.envoyproxy.io" 172.18.255.200/s2
    ```

    (Should route to echoserver-2)

    ```shell
    curl -i -H "Host: gateway.envoyproxy.io" 172.18.255.200/any-other-prefix
    ```

    (Should route to echoserver-3)
