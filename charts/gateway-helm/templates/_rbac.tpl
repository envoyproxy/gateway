{{/*
All namespaced resources for Envoy Gateway RBAC.
*/}}
{{- define "eg.rbac.namespaced" -}}
- {{ include "eg.rbac.namespaced.basic" . | nindent 2 | trim }}
- {{ include "eg.rbac.namespaced.apps" . | nindent 2 | trim }}
- {{ include "eg.rbac.namespaced.discovery" . | nindent 2 | trim }}
- {{ include "eg.rbac.namespaced.config.envoyproxy" . | nindent 2 | trim }}
- {{ include "eg.rbac.namespaced.gateway.envoyproxy" . | nindent 2 | trim }}
- {{ include "eg.rbac.namespaced.gateway.networking" . | nindent 2 | trim }}
- {{ include "eg.rbac.namespaced.gateway.networking.status" . | nindent 2 | trim }}
{{- end }}

{{/*
All cluster scoped resources for Envoy Gateway RBAC.
*/}}
{{- define "eg.rbac.cluster" -}}
- {{ include "eg.rbac.cluster.basic" . | nindent 2 | trim }}
- {{ include "eg.rbac.cluster.gateway.networking" . | nindent 2 | trim }}
- {{ include "eg.rbac.cluster.gateway.networking.status" . | nindent 2 | trim }}
{{- end }}

{{/*
Namespaced
*/}}

{{- define "eg.rbac.namespaced.basic" -}}
apiGroups:
- ""
resources:
- secrets
- services
verbs:
- get
- list
- watch
{{- end }}

{{- define "eg.rbac.namespaced.apps" -}}
apiGroups:
- apps
resources:
- deployments
verbs:
- get
- list
- watch
{{- end }}

{{- define "eg.rbac.namespaced.discovery" -}}
apiGroups:
- discovery.k8s.io
resources:
- endpointslices
verbs:
- get
- list
- watch
{{- end }}

{{- define "eg.rbac.namespaced.config.envoyproxy" -}}
apiGroups:
- config.gateway.envoyproxy.io
resources:
- envoyproxies
verbs:
- get
- list
- update
- watch
{{- end }}

{{- define "eg.rbac.namespaced.gateway.envoyproxy" -}}
apiGroups:
- gateway.envoyproxy.io
resources:
- authenticationfilters
- envoypatchpolicies
- ratelimitfilters
verbs:
- get
- list
- update
- watch
{{- end }}

{{- define "eg.rbac.namespaced.gateway.networking" -}}
apiGroups:
- gateway.networking.k8s.io
resources:
- gateways
- grpcroutes
- httproutes
- referencegrants
- referencepolicies
- tcproutes
- tlsroutes
- udproutes
verbs:
- get
- list
- patch
- update
- watch
{{- end }}

{{- define "eg.rbac.namespaced.gateway.networking.status" -}}
apiGroups:
- gateway.networking.k8s.io
resources:
- gateways/status
- grpcroutes/status
- httproutes/status
- tcproutes/status
- tlsroutes/status
- udproutes/status
verbs:
- update
{{- end }}

{{/*
Cluster scope
*/}}

{{- define "eg.rbac.cluster.basic" -}}
apiGroups:
- ""
resources:
- nodes
- namespaces
verbs:
- get
- list
- watch
{{- end }}

{{- define "eg.rbac.cluster.gateway.networking" -}}
apiGroups:
- gateway.networking.k8s.io
resources:
- gatewayclasses
verbs:
- get
- list
- patch
- update
- watch
{{- end }}

{{- define "eg.rbac.cluster.gateway.networking.status" -}}
apiGroups:
- gateway.networking.k8s.io
resources:
- gatewayclasses/status
verbs:
- update
{{- end }}
