{{- define "type" -}}
{{- $type := . -}}
{{- if markdownShouldRenderType $type -}}

#### {{ $type.Name }}

{{ $metaList := index .Markers "kubebuilder:metadata" }}
{{- range $meta := $metaList -}}
{{- range $anno := $meta.Annotations -}}
{{- if hasPrefix "gateway.envoyproxy.io/release-channel" $anno -}}
_Release Channel:_ {{ trimPrefix "gateway.envoyproxy.io/release-channel=" $anno }}
{{- end -}}
{{- end -}}
{{- end -}}

{{ if $type.IsAlias }}_Underlying type:_ _{{ markdownRenderTypeLink $type.UnderlyingType  }}_{{ end }}

{{ $type.Doc }}

{{ if $type.References -}}
_Appears in:_
{{- range $type.SortedReferences }}
- {{ markdownRenderTypeLink . }}
{{- end }}
{{- end }}

{{ if $type.Members -}}
| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
{{ if $type.GVK -}}
| `apiVersion` | _string_ | |`{{ $type.GVK.Group }}/{{ $type.GVK.Version }}`
| `kind` | _string_ | |`{{ $type.GVK.Kind }}`
{{ end -}}

{{ range $type.Members -}}
| `{{ .Name  }}` | _{{ markdownRenderType .Type }}_ | {{ with .Markers.optional }} {{ "false" }} {{ else }} {{ "true" }} {{end}} | {{ template "type_members" . }} |
{{ end -}}

{{ end -}}

{{- end -}}
{{- end -}}
