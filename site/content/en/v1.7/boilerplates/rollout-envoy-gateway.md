---
---

> After updating the `ConfigMap`, you will need to wait the configuration kicks in. <br/>
> You can **force** the configuration to be reloaded by restarting the `envoy-gateway` deployment.
>
> ```shell
> kubectl rollout restart deployment envoy-gateway -n envoy-gateway-system
> ```
>