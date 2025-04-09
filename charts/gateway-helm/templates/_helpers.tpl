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
{{- $registryName := default .Values.deployment.envoyGateway.image.registry ((.global).imageRegistry) -}}
{{- $repositoryName := .Values.deployment.envoyGateway.image.repository -}}
{{- $termination := .Values.deployment.envoyGateway.image.tag -}}
{{- printf "%s/%s:%s" $registryName $repositoryName $termination -}}
{{- end -}}

{{/*
Pull policy for the Envoy Gateway image.
*/}}
{{- define "eg.image.pullPolicy" -}}
{{ .Values.deployment.envoyGateway.imagePullPolicy }}
{{- end }}

{{/*
Pull secrets for the Envoy Gateway image.
*/}}
{{- define "eg.image.pullSecrets" -}}
{{- if .Values.global.imagePullSecrets -}}
{{ toYaml .Values.global.imagePullSecrets }}
{{- else if .Values.deployment.envoyGateway.imagePullSecrets -}}
{{ toYaml .Values.deployment.envoyGateway.imagePullSecrets }}
{{- else -}}
[]
{{- end }}
{{- end }}

{{/*
The name of the Envoy Ratelimit image.
*/}}
{{- define "eg.ratelimit.image" -}}
{{- $registryName := default .Values.ratelimit.image.registry ((.global).imageRegistry) -}}
{{- $repositoryName := .Values.ratelimit.image.repository -}}
{{- $termination := .Values.ratelimit.image.tag -}}
{{- printf "%s/%s:%s" $registryName $repositoryName $termination -}}
{{- end -}}

{{/*
Pull secrets for the Envoy Ratelimit image.
*/}}
{{- define "eg.ratelimit.image.pullSecrets" -}}
{{- if .Values.global.imagePullSecrets -}}
{{ toYaml .Values.global.imagePullSecrets }}
{{- else if .Values.ratelimit.imagePullSecrets -}}
{{ toYaml .Values.ratelimit.imagePullSecrets }}
{{- else -}}
{{ toYaml list }}
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
        image: {{ include "eg.ratelimit.image" . }}
      pod:
        imagePullSecrets: {{- include "eg.ratelimit.image.pullSecrets" . | nindent 10 }}
      patch:
        type: StrategicMerge
        value:
          spec:
            template:
              spec:
                containers:
                - name: envoy-ratelimit
                  imagePullPolicy: {{ .Values.ratelimit.imagePullPolicy }}
    shutdownManager:
      image: {{ include "eg.image" . }}
{{- end }}
