{{- define "gvList" -}}
{{- $groupVersions := . -}}

---
title: "API Reference"
aliases: "/api/extension_types"
---


## Packages
{{- range $groupVersions }}
- {{ markdownRenderGVLink . }}
{{- end }}

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}
