{{/*
Expand the name of the chart.
*/}}
{{- define "eg.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "eg.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "eg.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "eg.labels" -}}
helm.sh/chart: {{ include "eg.chart" . }}
{{ include "eg.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "eg.selectorLabels" -}}
app.kubernetes.io/name: {{ include "eg.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "eg.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "eg.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
The name of the Envoy Gateway image.
*/}}
{{- define "eg.image" -}}
{{-   $registry := .Values.deployment.envoyGateway.image.registry | default .Values.global.image.registry -}}
{{-   $repository := .Values.deployment.envoyGateway.image.repository -}}
{{-   $tag := .Values.deployment.envoyGateway.image.tag | default .Chart.AppVersion -}}
{{- printf "%s/%s:%s" $registry $repository $tag -}}
{{- end -}}

{{/*
The name of the Envoy Ratelimit image.
*/}}
{{- define "eg.ratelimit.image" -}}
{{-   $registry := .Values.ratelimit.image.registry | default .Values.global.image.registry -}}
{{-   $repository := .Values.ratelimit.image.repository -}}
{{-   $tag := .Values.ratelimit.image.tag -}}
{{-   printf "%s/%s:%s" $registry $repository $tag -}}
{{- end -}}

{{/*
Render imagePullSecrets conditionally.
If empty, renders as a single-line empty list.
If not empty, renders as a multi-line YAML list.
*/}}
{{- define "eg.imagePullSecrets" -}}
{{- $pullSecrets := .Values.deployment.envoyGateway.image.pullSecrets | default .Values.global.image.pullSecrets -}}
{{- if $pullSecrets }}
imagePullSecrets:
  {{- toYaml $pullSecrets | nindent 2 }}
{{- else }}
imagePullSecrets: []
{{- end -}}
{{- end -}}


{{/*
The default Envoy Gateway configuration.
*/}}
{{- define "eg.default-envoy-gateway-config" -}}
provider:
  type: Kubernetes
  kubernetes:
    rateLimitDeployment:
      container:
        image: {{ include "eg.ratelimit.image" . }}
      {{- if (or .Values.global.image.pullSecrets .Values.ratelimit.image.pullSecrets) }}
      pod:
        imagePullSecrets: {{ .Values.ratelimit.image.pullSecrets | default .Values.global.image.pullSecrets | toYaml | nindent 10 }}
      {{- end }}
      {{- if (or .Values.global.image.pullPolicy .Values.ratelimit.image.pullPolicy) }}
      patch:
        type: StrategicMerge
        value:
          spec:
            template:
              spec:
                containers:
                - name: envoy-ratelimit
                  imagePullPolicy: {{ .Values.ratelimit.image.pullPolicy | default .Values.global.image.pullPolicy }}
      {{- end }}
    shutdownManager:
      image: {{ include "eg.image" . }}
{{- with .Values.config.envoyGateway.extensionApis }}
extensionApis:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- if not .Values.topologyInjector.enabled }}
proxyTopologyInjector:
  disabled: true
{{- end }}
{{- end }}
