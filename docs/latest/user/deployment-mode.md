### One GatewayClass per Envoy Gateway

* Envoy Gateway can accept a single [GatewayClass](https://gateway-api.sigs.k8s.io/api-types/gatewayclass/)
resource. If you've instantiated multiple GatewayClasses, we recommend running multiple Envoy Gateway controllers
in different namespaces, linking a GatewayClass to each of them. 
* Support for accepting multiple GatewayClass is being tracked [here](https://github.com/envoyproxy/gateway/issues/1231).

### Deployment Mode

* The current deployment model is - Envoy Gateway **watches** for resources such a `Service` & `HTTPRoute` in **all** namespaces
and **creates** managed data plane resources such as EnvoyProxy `Deployment` in the **namespace where Envoy Gateway is running**.
* Support for alternate deployment modes is being tracked [here](https://github.com/envoyproxy/gateway/issues/1117).
