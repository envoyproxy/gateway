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


{{/*
Create the name of the webhook service
*/}}
{{- define "eg.webhookService" -}}
{{- printf "%s-webhook-service" (include "eg.name" .) -}}
{{- end -}}

{{/*
Create the name of the webhook cert secret
*/}}
{{- define "eg.webhookCertSecret" -}}
{{- printf "%s-webhook-tls" (include "eg.name" .) -}}
{{- end -}}

{{/*
Generate certificates for webhook
*/}}
{{- define "eg.webhookCerts" -}}
{{- $serviceName := (include "eg.webhookService" .) -}}
{{- $secretName := (include "eg.webhookCertSecret" .) -}}
{{- $secret := lookup "v1" "Secret" .Release.Namespace $secretName -}}
{{- if (and .Values.topologyWebhook.tls.caCert .Values.topologyWebhook.tls.cert .Values.topologyWebhook.tls.key) -}}
caCert: {{ .Values.topologyWebhook.tls.caCert | b64enc }}
clientCert: {{ .Values.topologyWebhook.tls.cert | b64enc }}
clientKey: {{ .Values.topologyWebhook.tls.key | b64enc }}
{{- else if and .Values.topologyWebhook.keepTLSSecret $secret -}}
caCert: {{ index $secret.data "ca.crt" }}
clientCert: {{ index $secret.data "tls.crt" }}
clientKey: {{ index $secret.data "tls.key" }}
{{- else -}}
{{- $altNames := list (printf "%s.%s" $serviceName .Release.Namespace) (printf "%s.%s.svc" $serviceName .Release.Namespace) (printf "%s.%s.svc.%s" $serviceName .Release.Namespace .Values.topologyWebhook.clusterDnsDomain) -}}
{{- $ca := genCA "eg-ca" 3650 -}}
{{- $cert := genSignedCert (include "eg.fullname" .) nil $altNames 3650 $ca -}}
caCert: {{ $ca.Cert | b64enc }}
clientCert: {{ $cert.Cert | b64enc }}
clientKey: {{ $cert.Key | b64enc }}
{{- end -}}
{{- end -}}


