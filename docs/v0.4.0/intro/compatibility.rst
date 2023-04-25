Compatibility Matrix
====================

Envoy Gateway relies on the Envoy Proxy and the Gateway API, and runs
within a Kubernetes cluster. Not all versions of each of these products
can function together for Envoy Gateway. Supported version combinations
are listed below; **bold** type indicates the versions of the Envoy Proxy
and the Gateway API actually compiled into each Envoy Gateway release.

+--------------------------+---------------------+---------------------+---------------------+----------------------------+
| Envoy Gateway version    | Envoy Proxy version | Rate Limit version  | Gateway API version | Kubernetes version         |
+--------------------------+---------------------+---------------------+---------------------+----------------------------+
| v0.4.0                   | **v1.26-latest**    |   **542a6047**      |  **v0.6.1**         | v1.25, v1.26, v1.27        |
+--------------------------+---------------------+---------------------+---------------------+----------------------------+
| v0.3.0                   | **v1.25-latest**    |   **f28024e3**      |  **v0.6.1**         | v1.24, v1.25, v1.26        |
+--------------------------+---------------------+---------------------+---------------------+----------------------------+
| v0.2.0                   | **v1.23-latest**    |                     |  **v0.5.1**         | v1.24                      |
+--------------------------+---------------------+---------------------+---------------------+----------------------------+
| latest                   | **dev-latest**      |   **master**        |  **v0.6.2**         | v1.25, v1.26, v1.27        |
+--------------------------+---------------------+---------------------+---------------------+----------------------------+

.. note::

   This project is under active development. Many, many features are not
   complete. We would love for you to :doc:`get involved<../get_involved>`.
