{{- define "gvList" -}}
{{- $groupVersions := . -}}

+++
title = "Gateway API Extensions"
weight = 1
description = "Envoy Gateway provides these extensions to support additional features not available in the Gateway API today"
+++


## Packages
{{- range $groupVersions }}
- {{ markdownRenderGVLink . }}
{{- end }}

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}
