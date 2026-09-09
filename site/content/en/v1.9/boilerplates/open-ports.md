## Open Ports

These are the ports used by Envoy Gateway and the managed Envoy Proxy.

### Envoy Gateway

|     Envoy Gateway     |  Address  | Port  | Configurable |
| :-------------------: | :-------: | :---: | :----------: |
| Xds EnvoyProxy Server |  0.0.0.0  | 18000 |      No      |
| Xds RateLimit Server  |  0.0.0.0  | 18001 |      No      |
|     Admin Server      | 127.0.0.1 | 19000 |     Yes      |
|    Metrics Server     |  0.0.0.0  | 19001 |      No      |
|     Health Check      | 127.0.0.1 | 8081  |      No      |

### EnvoyProxy

|   Envoy Proxy    |  Address  | Port  |
| :--------------: | :-------: | :---: |
|   Admin Server   | 127.0.0.1 | 19000 |
|      Stats       |  0.0.0.0  | 19001 |
| Shutdown Manager |  0.0.0.0  | 19002 |
|    Readiness     |  0.0.0.0  | 19003 |
