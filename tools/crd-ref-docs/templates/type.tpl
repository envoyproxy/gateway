{{- define "type" -}}
{{- $type := . -}}
{{- if markdownShouldRenderType $type -}}

#### {{ $type.Name }}

{{ if $type.IsAlias }}_Underlying type:_ _{{ markdownRenderTypeLink $type.UnderlyingType  }}_{{ end }}

{{ $type.Doc }}

{{ if $type.References -}}
_Appears in:_
{{- range $type.SortedReferences }}
- {{ markdownRenderTypeLink . }}
{{- end }}
{{- end }}

{{ if $type.Members -}}
| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
{{ if $type.GVK -}}
| `apiVersion` | _string_ | |`{{ $type.GVK.Group }}/{{ $type.GVK.Version }}`
| `kind` | _string_ | |`{{ $type.GVK.Kind }}`
{{ end -}}

{{ range $type.Members -}}
{{- with .Markers.notImplementedHide -}}
{{- else -}}
| `{{ .Name  }}` | _{{ markdownRenderType .Type }}_ | {{ with .Markers.optional }} {{ "false" }} {{ else }} {{ "true" }} {{end}} | {{ markdownRenderDefault .Default }} | {{ template "type_members" . }} |
{{ end -}}
{{- end -}}
{{- end -}}
{{ if $type.EnumValues -}}
| Value | Description |
| ----- | ----------- |
{{ range $type.EnumValues -}}
| `{{ .Name }}` | {{ markdownRenderFieldDoc .Doc }} | 
{{ end -}}
{{- end -}}
{{- end -}}
{{- end -}}
