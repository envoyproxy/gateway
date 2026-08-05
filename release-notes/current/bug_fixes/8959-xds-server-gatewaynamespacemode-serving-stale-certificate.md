Fixed the xDS server in GatewayNamespaceMode serving a stale certificate after cert-manager rotation by re-reading the cert from disk on every TLS handshake.
