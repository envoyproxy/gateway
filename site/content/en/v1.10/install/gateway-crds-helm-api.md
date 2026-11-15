+++
title = "Gateway Crds Helm Chart"
+++

![Version: v0.0.0-latest](https://img.shields.io/badge/Version-v0.0.0--latest-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: latest](https://img.shields.io/badge/AppVersion-latest-informational?style=flat-square)

A Helm chart for Envoy Gateway CRDs

**Homepage:** <https://gateway.envoyproxy.io/>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| envoy-gateway-steering-committee |  | <https://github.com/envoyproxy/gateway/blob/main/GOVERNANCE.md> |
| envoy-gateway-maintainers |  | <https://github.com/envoyproxy/gateway/blob/main/CODEOWNERS> |

## Source Code

* <https://github.com/envoyproxy/gateway>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| crds.envoyGateway.enabled | bool | `false` |  |
| crds.gatewayAPI.channel | string | `"experimental"` |  |
| crds.gatewayAPI.enabled | bool | `false` |  |

