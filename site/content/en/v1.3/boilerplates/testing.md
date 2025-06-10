---

---

{{< tabpane text=true >}}
{{% tab header="With External LoadBalancer Support" %}}

Get the External IP of the Gateway:
```shell
export GATEWAY_HOST=$(kubectl get gateway/eg -o jsonpath='{.status.addresses[0].value}')
```

Curl the example app through Envoy proxy:

```shell
curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST/get
```

The above command should succeed with status code 200. 


{{% /tab %}}
{{% tab header="Without LoadBalancer Support" %}}

Get the name of the Envoy service created the by the example Gateway:

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

The above command should succeed with status code 200. 

{{% /tab %}}
{{< /tabpane >}}
