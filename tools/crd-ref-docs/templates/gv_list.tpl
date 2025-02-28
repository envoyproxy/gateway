{{- define "gvList" -}}
{{- $groupVersions := . -}}

+++
title = "Envoy Gateway API"
weight = 1
+++


## Packages
{{- range $groupVersions }}
- {{ markdownRenderGVLink . }}
{{- end }}

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}
