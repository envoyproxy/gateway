{{/*
Rule set granted to the optional aggregated `view` ClusterRole: read-only
access to the Gateway API and Envoy Gateway resources surfaced by this
chart. Reuses the same resource lists as the controller's own RBAC so the
two never drift apart.
*/}}
{{- define "eg.rbac.aggregated.view" -}}
- {{ include "eg.rbac.namespaced.gateway.networking" . | nindent 2 | trim }}
- apiGroups:
  - gateway.networking.k8s.io
  resources:
  - gatewayclasses
  verbs:
  - get
  - list
  - watch
- {{ include "eg.rbac.namespaced.gateway.envoyproxy" . | nindent 2 | trim }}
{{- end }}

{{/*
Rule set granted to the optional aggregated `edit`/`admin` ClusterRoles:
the same resource set as `eg.rbac.aggregated.view`, plus write verbs.
`edit` and `admin` carry identical rules for these resources; `admin`'s
additional elevation over `edit` is already supplied by the Kubernetes
built-in `admin` ClusterRole this one aggregates into.

EnvoyPatchPolicy, EnvoyProxy, GatewayClass, Backend,
EnvoyExtensionPolicy, and ReferenceGrant are deliberately excluded from
write access here. EnvoyPatchPolicy lets its holder inject arbitrary
proxy configuration, up to full compromise of the shared proxy (see the
Security Warning in
site/content/en/latest/tasks/extensibility/envoy-patch-policy.md).
EnvoyProxy's `envoyDeployment.patch` field is honoured by default (the
`EnvoyProxyPatch` runtime flag) and lets its holder merge-patch the proxy
Deployment applied by Envoy Gateway's more privileged service account,
up to injecting `hostPath`/`hostNetwork`/privileged containers and
escalating from namespace-scoped access to node-root and cluster
compromise (see site/content/en/latest/tasks/security/threat-model.md).
GatewayClass is cluster-scoped, not namespace-scoped like the rest of
this rule set; the project's threat model limits Application Admin write
access to Gateways within their own namespaces, reserving GatewayClass
writes for the cluster operator tier. Backend can bypass ReferenceGrants
to reach otherwise restricted Services or Pods, or expose the Envoy
admin endpoint, and the project's own docs advise admins to restrict it
with RBAC (see site/content/en/latest/tasks/traffic/backend.md).
EnvoyExtensionPolicy can carry Lua scripts that run inside the proxy
process without strong sandboxing, and the project's own docs tell
admins to restrict who can create or modify it (see
site/content/en/latest/tasks/extensibility/lua.md). ReferenceGrant
authorizes cross-namespace trust, and the project's threat model
(EGTM-012) says the ability to create it should be restricted to the
minimum number of people with special attention paid to ClusterRoles.
Granting any of these through a generic edit/admin aggregation would
hand that capability to any subject bound to those built-in roles, which
the project's threat model and docs explicitly advise against.
*/}}
{{- define "eg.rbac.aggregated.write" -}}
- apiGroups:
  - gateway.networking.k8s.io
  resources:
  - gateways
  - listenersets
  - grpcroutes
  - httproutes
  - tcproutes
  - tlsroutes
  - udproutes
  - backendtlspolicies
  verbs:
  - create
  - update
  - patch
  - delete
  - deletecollection
- apiGroups:
  - gateway.envoyproxy.io
  resources:
  - clienttrafficpolicies
  - backendtrafficpolicies
  - securitypolicies
  - httproutefilters
  verbs:
  - create
  - update
  - patch
  - delete
  - deletecollection
{{- end }}

{{/*
All namespaced resources for Envoy Gateway RBAC.
*/}}
{{- define "eg.rbac.namespaced" -}}
- {{ include "eg.rbac.namespaced.basic" . | nindent 2 | trim }}
- {{ include "eg.rbac.namespaced.apps" . | nindent 2 | trim }}
- {{ include "eg.rbac.namespaced.discovery" . | nindent 2 | trim }}
- {{ include "eg.rbac.namespaced.gateway.envoyproxy" . | nindent 2 | trim }}
- {{ include "eg.rbac.namespaced.gateway.envoyproxy.status" . | nindent 2 | trim }}
- {{ include "eg.rbac.namespaced.gateway.networking" . | nindent 2 | trim }}
- {{ include "eg.rbac.namespaced.gateway.networking.status" . | nindent 2 | trim }}
{{- if .Values.topologyInjector.enabled }}
- {{ include "eg.rbac.namespaced.topologyinjector" . | nindent 2 | trim }}
{{- end }}
{{- end }}

{{/*
All cluster scoped resources for Envoy Gateway RBAC.
*/}}
{{- define "eg.rbac.cluster" -}}
- {{ include "eg.rbac.cluster.basic" . | nindent 2 | trim }}
- {{ include "eg.rbac.cluster.gateway.networking" . | nindent 2 | trim }}
- {{ include "eg.rbac.cluster.gateway.networking.status" . | nindent 2 | trim }}
- {{ include "eg.rbac.cluster.multiclusterservices" . | nindent 2 | trim }}
{{- end }}

{{/*
Namespaced
*/}}

{{- define "eg.rbac.namespaced.basic" -}}
apiGroups:
- ""
resources:
- configmaps
- secrets
- services
verbs:
- get
- list
- watch
{{- end }}

{{- define "eg.rbac.namespaced.topologyinjector" -}}
apiGroups:
- ""
resources:
- pods
- pods/binding
verbs:
- get
- list
- patch
- update
- watch
{{- end }}

{{- define "eg.rbac.namespaced.apps" -}}
apiGroups:
- apps
resources:
- deployments
- daemonsets
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

{{- define "eg.rbac.namespaced.gateway.envoyproxy" -}}
apiGroups:
- gateway.envoyproxy.io
resources:
- envoyproxies
- envoypatchpolicies
- clienttrafficpolicies
- backendtrafficpolicies
- securitypolicies
- envoyextensionpolicies
- backends
- httproutefilters
verbs:
- get
- list
- watch
{{- end }}

{{- define "eg.rbac.namespaced.gateway.envoyproxy.status" -}}
apiGroups:
- gateway.envoyproxy.io
resources:
- envoyproxies/status
- envoypatchpolicies/status
- clienttrafficpolicies/status
- backendtrafficpolicies/status
- securitypolicies/status
- envoyextensionpolicies/status
- backends/status
verbs:
- update
{{- end }}

{{- define "eg.rbac.namespaced.gateway.networking" -}}
apiGroups:
- gateway.networking.k8s.io
resources:
- gateways
- listenersets
- grpcroutes
- httproutes
- referencegrants
- tcproutes
- tlsroutes
- udproutes
- backendtlspolicies
verbs:
- get
- list
- watch
{{- end }}

{{- define "eg.rbac.namespaced.gateway.networking.status" -}}
apiGroups:
- gateway.networking.k8s.io
resources:
- gateways/status
- listenersets/status
- grpcroutes/status
- httproutes/status
- tcproutes/status
- tlsroutes/status
- udproutes/status
- backendtlspolicies/status
verbs:
- update
{{- end }}

{{- define "eg.rbac.namespaced.gateway.networking.experimental" -}}
{{- end }}

{{- define "eg.rbac.namespaced.gateway.networking.experimental.status" -}}
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


{{- define "eg.rbac.cluster.multiclusterservices" -}}
apiGroups:
- multicluster.x-k8s.io
resources:
- serviceimports
verbs:
- get
- list
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

{{- define "eg.rbac.infra.basic" -}}
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  - services
  - configmaps
  verbs:
  - create
  - get
  - list
  - delete
  - deletecollection
  - patch
  - watch
- apiGroups:
  - apps
  resources:
  - deployments
  - daemonsets
  verbs:
  - create
  - get
  - list
  - delete
  - deletecollection
  - patch
  - watch
- apiGroups:
  - autoscaling
  resources:
  - horizontalpodautoscalers
  verbs:
  - create
  - get
  - list
  - delete
  - deletecollection
  - patch
  - watch
- apiGroups:
  - policy
  resources:
  - poddisruptionbudgets
  verbs:
  - create
  - get
  - list
  - delete
  - deletecollection
  - patch
  - watch
- apiGroups:
  - certificates.k8s.io
  resources:
  - clustertrustbundles
  verbs:
  - list
  - get
  - watch
{{- end }}

{{- define "eg.rbac.controllernamespace.secrets.read" -}}
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - list
  - watch
{{- end }}

{{- define "eg.rbac.infra.tokenreview" -}}
- apiGroups:
  - authentication.k8s.io
  resources:
  - tokenreviews
  verbs:
  - create
{{- end }}
