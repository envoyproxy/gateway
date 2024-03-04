echo "Apply Envoy gateway configurations"
kubectl apply -f test/config/gatewayclass.yaml
kubectl apply -f test/benchmark/gateway.yaml

HTTPROUTE_MUM=$HTTPROUTE_MUM1:-1
for i in $(seq 1 $HTTPROUTE_MUM); do
    # replace the httproute name
    old_name=benchmark-test-server-$i
    new_name=benchmark-test-server-$(expr $i + 1)
    sed -i `s/benchmark-test-server-$i/g`
    echo "Applying HTTPROUTE $new_name"
    kubectl apply -f test/benchmark/httproute.yaml
done
kubectl wait --timeout=$WAIT_TIMEOUT -n envoy-gateway-system pods -l gateway.envoyproxy.io/owning-gateway-name=benchmark --for condition=Ready 

ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system -l gateway.envoyproxy.io/owning-gateway-name=benchmark -o jsonpath='{.items[0].metadata.name}')
NODE_HOST=$(kubectl get node -o jsonpath='{.items[0].status.addresses[0].address}')
kubectl port-forward -n envoy-gateway-system service/$ENVOY_SERVICE 8081:8081 &

echo "Running benchmark tests"
docker run envoyproxy/nighthawk-dev:latest nighthawk_client --rps $RPS --connections $CONNECTIONS --request-header "Host: www.example.com" --duration $DURATION --concurrency auto http://$NODE_HOST:8081
