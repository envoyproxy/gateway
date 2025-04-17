+++
title = "Advanced: Envoy Proxy Admin Interface"
+++

## Overview
Platform admins looking to troubleshoot low level aspects of the data plane such as xDS config and heap dump, can directly connect to the Envoy Proxy Admin Interface.

**Note**: Application Developers may not have access to the namespace where the Envoy Proxy fleet is running and should rely on [exported telemetry](https://gateway.envoyproxy.io/docs/tasks/observability/) instead for troubleshooting.

## Prerequisites

{{< boilerplate prerequisites >}}

### Access

You will need to port-forward to the admin interface port (currently 19000) on the Envoy deployment that corresponds to a Gateway, since it only listens on the `localhost`
address for security reasons.

Get the name of the Envoy deployment. In this example its for Gateway `eg` in the `default` namespace:

```shell
export ENVOY_DEPLOYMENT=$(kubectl get deploy -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
```

Port forward to it.

```shell
kubectl port-forward deploy/${ENVOY_DEPLOYMENT} -n envoy-gateway-system 19000:19000 &
```

If you enter `http://localhost:19000` in a browser, you should be able to access the admin interface.


Here's another example of accessing the `/config_dump` endpoint to get access of the programmed xDS configuration.

```shell
curl http://127.0.0.1:19000/config_dump
```

### Next Steps

There are many other endpoints in the [Envoy Proxy Admin interface](https://www.envoyproxy.io/docs/envoy/latest/operations/admin) that may be helpful when debugging.
