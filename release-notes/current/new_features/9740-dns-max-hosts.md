Added a `maxHosts` field to the DNS configuration, exposing Envoy's
`DnsCacheConfig.max_hosts` for the dynamic forward proxy DNS cache. The field is
available everywhere existing DNS settings are configurable
(BackendTrafficPolicy, EnvoyExtensionPolicy, EnvoyProxy, and SecurityPolicy) and
bounds how many hosts the DNS cache holds. When unset, Envoy's default of 1024 is
preserved.
