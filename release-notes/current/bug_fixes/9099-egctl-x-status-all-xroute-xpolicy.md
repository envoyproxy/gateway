Fixed `egctl x status all`/`xroute`/`xpolicy` failing when a Gateway API CRD (e.g. TCPRoute) is not installed in the cluster; missing CRDs are now skipped silently, or reported on stderr with `-v`.
