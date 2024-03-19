echo "-- Deploying benchmark test server ---"
kubectl create namespace nighthawk-test-server
kubectl -n nighthawk-test-server create configmap test-server-config --from-file=test/benchmark/test-server.yaml --output yaml
kubectl apply -f test/benchmark/benchmark-test-server.yaml
kubectl wait --timeout=$WAIT_TIMEOUT -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
kubectl wait --timeout=$WAIT_TIMEOUT -n nighthawk-test-server deployment/nighthawk-test-server --for=condition=Available

echo "-- Download the Promtool binary ---"
wget -qO- "https://github.com/prometheus/prometheus/releases/download/v2.50.1/prometheus-2.50.1.linux-amd64.tar.gz" | tar xvzf - "prometheus-2.50.1.linux-amd64"/promtool --strip-components=1
./promtool --version

echo "--- Apply Gateway API configurations ---"
kubectl apply -f test/config/gatewayclass.yaml
# TODO: Now it applies two gatewaies to the same backend server, we should make it configurable
kubectl apply -f test/benchmark/gateway.yaml
kubectl get gateway -A

HTTPROUTE_NUM=500
# TODO: Now it applies HTTPRoutes to the same gateway, we should make it configurable
echo "Number of HTTROUTE: $HTTPROUTE_NUM"
if [ $HTTPROUTE_NUM -eq 1 ]; then
    echo "Applying HTTPROUTE configuration"
    kubectl apply -f test/benchmark/httproute.yaml
else
    for i in $(seq 1 $HTTPROUTE_NUM); do
        if [ $i -eq 1 ]; then
            echo "Applying HTTPROUTE benchmark-test-server-1"
            echo "Applying HTTPROUTE benchmark-test-server-2"
            kubectl apply -f test/benchmark/httproute.yaml
        else
            # replace the httproute name
            old_name_1=benchmark-test-server-1-$(expr $i - 1)
            new_name_1=benchmark-test-server-1-$i
            old_name_2=benchmark-test-server-2-$(expr $i - 1)
            new_name_2=benchmark-test-server-2-$i
            echo "Applying HTTPROUTE $new_name_1"
            echo "Applying HTTPROUTE $new_name_2"
            sed -i "s/$old_name_1/$new_name_1/g" test/benchmark/httproute.yaml
            sed -i "s/$old_name_2/$new_name_2/g" test/benchmark/httproute.yaml
            kubectl apply -f test/benchmark/httproute.yaml
        fi
    done
fi
kubectl get httproute -A

echo "--- Wating for Envoy gateway data plane to be ready ---"
sleep 10
kubectl wait --timeout=$WAIT_TIMEOUT -n envoy-gateway-system pods -l gateway.envoyproxy.io/owning-gateway-namespace=nighthawk-test-server --for condition=Ready 

echo "--- Deploying Kubernetes Metrics server and Prometheus to collect metrics ---"
# https://gateway.envoyproxy.io/v0.6.0/user/gateway-api-metrics/
kubectl apply --server-side -f https://raw.githubusercontent.com/Kuadrant/gateway-api-state-metrics/main/config/examples/kube-prometheus/bundle_crd.yaml
kubectl apply -f https://raw.githubusercontent.com/Kuadrant/gateway-api-state-metrics/main/config/examples/kube-prometheus/bundle.yaml
sleep 60
kubectl -n monitoring rollout status --watch --timeout=20m statefulset/prometheus-k8s
kubectl -n monitoring port-forward service/prometheus-k8s 9090:9090 > /dev/null &

echo "--- Port-forwarding Envoy gateway data plane ---"
ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system -l gateway.envoyproxy.io/owning-gateway-namespace=nighthawk-test-server -o jsonpath='{.items[0].metadata.name}')
NODE_HOST=$(kubectl get node -o jsonpath='{.items[0].status.addresses[0].address}')
kubectl port-forward -n envoy-gateway-system service/$ENVOY_SERVICE 8081:8081 > /dev/null &

echo "--- Running benchmark tests ---"
docker run envoyproxy/nighthawk-dev:latest nighthawk_client --rps $RPS --connections $CONNECTIONS --request-header "Host: www.example.com" --duration $DURATION --concurrency auto http://$NODE_HOST:8081

echo "--- Collecting metrics ---"
ENVOY_CP_POD=$(kubectl get pods -n envoy-gateway-system -l control-plane=envoy-gateway -o jsonpath='{.items[0].metadata.name}')
./promtool query instant --time $DURATION http://localhost:9090/ "rate(container_cpu_usage_seconds_total{pod='$ENVOY_CP_POD'}[$DURITION])"

echo "--- Deleting Promtool ---"
rm -f ./promtool