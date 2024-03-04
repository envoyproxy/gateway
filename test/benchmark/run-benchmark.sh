echo "--- Apply Gateway API configurations ---"
kubectl apply -f test/config/gatewayclass.yaml
kubectl apply -f test/benchmark/gateway.yaml

echo "Number of HTTROUTE: $HTTPROUTE_NUM"
if [ $HTTPROUTE_NUM -eq 1 ]; then
    echo "Applying HTTPROUTE configuration"
    kubectl apply -f test/benchmark/httproute.yaml
else
    for i in $(seq 1 $HTTPROUTE_NUM); do
        echo "Applying HTTPROUTE $new_name"
        kubectl apply -f test/benchmark/httproute.yaml
        # replace the httproute name
        old_name=benchmark-test-server-$i
        new_name=benchmark-test-server-$(expr $i + 1)
        sed -i "s/$old_name/$new_name/g" test/benchmark/httproute.yaml
    done
fi

echo "--- Wating for Envoy gateway data plane to be ready ---"
sleep 10
kubectl wait --timeout=$WAIT_TIMEOUT -n envoy-gateway-system deployment --all --for=condition=Available

echo "--- Port-forwarding Envoy gateway data plane ---"
ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system -l gateway.envoyproxy.io/owning-gateway-name=benchmark -o jsonpath='{.items[0].metadata.name}')
NODE_HOST=$(kubectl get node -o jsonpath='{.items[0].status.addresses[0].address}')
kubectl port-forward -n envoy-gateway-system service/$ENVOY_SERVICE 8081:8081 &

echo "Running benchmark tests"
docker run envoyproxy/nighthawk-dev:latest nighthawk_client --rps $RPS --connections $CONNECTIONS --request-header "Host: www.example.com" --duration $DURATION --concurrency auto http://$NODE_HOST:8081
