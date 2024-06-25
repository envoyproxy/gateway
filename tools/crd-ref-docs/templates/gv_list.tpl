{{- define "gvList" -}}
{{- $groupVersions := . -}}

+++
title = "API Reference"
+++


## Packages
{{- range $groupVersions }}
- {{ markdownRenderGVLink . }}
{{- end }}

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}
