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
{{- if .Values.deployment.envoyGateway.image.repository }}
{{- .Values.deployment.envoyGateway.image.repository }}:{{ .Values.deployment.envoyGateway.image.tag | default .Values.global.images.envoyGateway.tag | default .Chart.AppVersion }}
{{- else if .Values.global.images.envoyGateway.image }}
{{- .Values.global.images.envoyGateway.image }}
{{- else }}
docker.io/envoyproxy/gateway:{{ .Chart.Version }}
{{- end }}
{{- end }}

{{/*
Pull policy for the Envoy Gateway image.
*/}}
{{- define "eg.image.pullPolicy" -}}
{{ .Values.deployment.envoyGateway.imagePullPolicy | default .Values.global.images.envoyGateway.pullPolicy | default "IfNotPresent" }}
{{- end }}

{{/*
Pull secrets for the Envoy Gateway image.
*/}}
{{- define "eg.image.pullSecrets" -}}
{{- if .Values.deployment.envoyGateway.imagePullSecrets -}}
imagePullSecrets:
{{ toYaml .Values.deployment.envoyGateway.imagePullSecrets }}
{{- else if .Values.global.images.envoyGateway.pullSecrets -}}
imagePullSecrets:
{{ toYaml .Values.global.images.envoyGateway.pullSecrets }}
{{- else -}}
imagePullSecrets: []
{{- end }}
{{- end }}

{{/*
The default Envoy Gateway configuration.
*/}}
{{- define "eg.default-envoy-gateway-config" -}}
provider:
  type: Kubernetes
  kubernetes:
    rateLimitDeployment:
      container:
        {{- if .Values.global.images.ratelimit.image }}
        image: {{ .Values.global.images.ratelimit.image }}
        {{- else }}
        image: "docker.io/envoyproxy/ratelimit:master"
        {{- end }}
      {{- with .Values.global.images.ratelimit.pullSecrets }}
      pod:
        imagePullSecrets:
        {{- toYaml . | nindent 10 }}
      {{- end }}
      {{- with .Values.global.images.ratelimit.pullPolicy }}
      patch:
        type: StrategicMerge
        value:
          spec:
            template:
              spec:
                containers:
                - name: envoy-ratelimit
                  imagePullPolicy: {{ . }}
      {{- end }}
    shutdownManager:
      image: {{ include "eg.image" . }}
{{- end }}
