---
title: "Load Balancing"
---
## Overview
Load balancing is the process of distributing incoming requests across multiple backend services to ensure efficient use of resources, high availability, and consistent performance. Rather than sending all traffic to a single backend—which can lead to slowdowns or outages—load balancing helps spread the load, making applications more resilient and scalable.

## Use Cases
Use load balancing to:
- Handle more traffic by spreading it across multiple service instances.
- Keep services available even if one backend goes down.
- Improve speed by routing requests to less-busy services.
- Keep user sessions sticky (e.g., always send a returning user to the same backend).


## Load Balancing in Envoy Gateway
Envoy Gateway supports the following load balancing strategies:
- **Round Robin:** Distributes requests sequentially to all backends.
- **Least Connections:** Routes requests to the backend with the fewest active connections.
- **Random:** Selects a backend randomly to balance traffic.
- **Consistent Hashing:** Distributes traffic based on hash values, ensuring requests from the same client go to the same backend (useful for stateful services).

## Related Resources
[Set up a Load Balancer](../tasks/traffic/load-balancing.md)