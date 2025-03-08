# gateway-addons-helm

![Version: v0.0.0-latest](https://img.shields.io/badge/Version-v0.0.0--latest-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: latest](https://img.shields.io/badge/AppVersion-latest-informational?style=flat-square)

An Add-ons Helm chart for Envoy Gateway

**Homepage:** <https://gateway.envoyproxy.io/>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| envoy-gateway-steering-committee |  | <https://github.com/envoyproxy/gateway/blob/main/GOVERNANCE.md> |
| envoy-gateway-maintainers |  | <https://github.com/envoyproxy/gateway/blob/main/CODEOWNERS> |

## Source Code

* <https://github.com/envoyproxy/gateway>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://fluent.github.io/helm-charts | fluent-bit | 0.30.4 |
| https://grafana.github.io/helm-charts | alloy | 0.9.2 |
| https://grafana.github.io/helm-charts | grafana | 8.0.0 |
| https://grafana.github.io/helm-charts | loki | 4.8.0 |
| https://grafana.github.io/helm-charts | tempo | 1.3.1 |
| https://open-telemetry.github.io/opentelemetry-helm-charts | opentelemetry-collector | 0.117.3 |
| https://prometheus-community.github.io/helm-charts | prometheus | 25.21.0 |

## Usage

[Helm](https://helm.sh) must be installed to use the charts.
Please refer to Helm's [documentation](https://helm.sh/docs) to get started.

The Envoy Gateway must be installed before installing this chart.

### Install from DockerHub

Once Helm has been set up correctly, install the chart from dockerhub:

``` shell
    helm install eg-addons oci://docker.io/envoyproxy/gateway-addons-helm --version v0.0.0-latest -n monitoring --create-namespace
```

You can find all helm chart release in [Dockerhub](https://hub.docker.com/r/envoyproxy/gateway-addons-helm/tags)

To uninstall the chart:

``` shell
    helm uninstall eg-addons -n monitoring
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| alloy.alloy.configMap.content | string | `"// Write your Alloy config here:\nlogging {\n  level = \"info\"\n  format = \"logfmt\"\n}\nloki.write \"alloy\" {\n  endpoint {\n    url = \"http://loki.monitoring.svc:3100/loki/api/v1/push\"\n  }\n}\n// discovery.kubernetes allows you to find scrape targets from Kubernetes resources.\n// It watches cluster state and ensures targets are continually synced with what is currently running in your cluster.\ndiscovery.kubernetes \"pod\" {\n  role = \"pod\"\n}\n\n// discovery.relabel rewrites the label set of the input targets by applying one or more relabeling rules.\n// If no rules are defined, then the input targets are exported as-is.\ndiscovery.relabel \"pod_logs\" {\n  targets = discovery.kubernetes.pod.targets\n\n  // Label creation - \"namespace\" field from \"__meta_kubernetes_namespace\"\n  rule {\n    source_labels = [\"__meta_kubernetes_namespace\"]\n    action = \"replace\"\n    target_label = \"namespace\"\n  }\n\n  // Label creation - \"pod\" field from \"__meta_kubernetes_pod_name\"\n  rule {\n    source_labels = [\"__meta_kubernetes_pod_name\"]\n    action = \"replace\"\n    target_label = \"pod\"\n  }\n\n  // Label creation - \"container\" field from \"__meta_kubernetes_pod_container_name\"\n  rule {\n    source_labels = [\"__meta_kubernetes_pod_container_name\"]\n    action = \"replace\"\n    target_label = \"container\"\n  }\n\n  // Label creation -  \"app\" field from \"__meta_kubernetes_pod_label_app_kubernetes_io_name\"\n  rule {\n    source_labels = [\"__meta_kubernetes_pod_label_app_kubernetes_io_name\"]\n    action = \"replace\"\n    target_label = \"app\"\n  }\n\n  // Label creation -  \"job\" field from \"__meta_kubernetes_namespace\" and \"__meta_kubernetes_pod_container_name\"\n  // Concatenate values __meta_kubernetes_namespace/__meta_kubernetes_pod_container_name\n  rule {\n    source_labels = [\"__meta_kubernetes_namespace\", \"__meta_kubernetes_pod_container_name\"]\n    action = \"replace\"\n    target_label = \"job\"\n    separator = \"/\"\n    replacement = \"$1\"\n  }\n\n  // Label creation - \"container\" field from \"__meta_kubernetes_pod_uid\" and \"__meta_kubernetes_pod_container_name\"\n  // Concatenate values __meta_kubernetes_pod_uid/__meta_kubernetes_pod_container_name.log\n  rule {\n    source_labels = [\"__meta_kubernetes_pod_uid\", \"__meta_kubernetes_pod_container_name\"]\n    action = \"replace\"\n    target_label = \"__path__\"\n    separator = \"/\"\n    replacement = \"/var/log/pods/*$1/*.log\"\n  }\n\n  // Label creation -  \"container_runtime\" field from \"__meta_kubernetes_pod_container_id\"\n  rule {\n    source_labels = [\"__meta_kubernetes_pod_container_id\"]\n    action = \"replace\"\n    target_label = \"container_runtime\"\n    regex = \"^(\\\\S+):\\\\/\\\\/.+$\"\n    replacement = \"$1\"\n  }\n}\n\n// loki.source.kubernetes tails logs from Kubernetes containers using the Kubernetes API.\nloki.source.kubernetes \"pod_logs\" {\n  targets    = discovery.relabel.pod_logs.output\n  forward_to = [loki.process.pod_logs.receiver]\n}\n// loki.process receives log entries from other Loki components, applies one or more processing stages,\n// and forwards the results to the list of receivers in the component’s arguments.\nloki.process \"pod_logs\" {\n  stage.static_labels {\n      values = {\n        cluster = \"envoy-gateway\",\n      }\n  }\n\n  forward_to = [loki.write.alloy.receiver]\n}"` |  |
| alloy.enabled | bool | `false` |  |
| alloy.fullnameOverride | string | `"alloy"` |  |
| dashboard.labels | object | `{}` |  |
| fluent-bit.config.filters | string | `"[FILTER]\n    Name kubernetes\n    Match kube.*\n    Merge_Log On\n    Keep_Log Off\n    K8S-Logging.Parser On\n    K8S-Logging.Exclude On\n\n[FILTER]\n    Name grep\n    Match kube.*\n    Regex $kubernetes['container_name'] ^envoy$\n\n[FILTER]\n    Name parser\n    Match kube.*\n    Key_Name log\n    Parser envoy\n    Reserve_Data True\n"` |  |
| fluent-bit.config.inputs | string | `"[INPUT]\n    Name tail\n    Path /var/log/containers/*.log\n    multiline.parser docker, cri\n    Tag kube.*\n    Mem_Buf_Limit 5MB\n    Skip_Long_Lines On\n"` |  |
| fluent-bit.config.outputs | string | `"[OUTPUT]\n    Name                   loki\n    Match                  kube.*\n    Host                   loki.monitoring.svc.cluster.local\n    Port                   3100\n    Labels                 job=fluentbit, app=$kubernetes['labels']['app'], k8s_namespace_name=$kubernetes['namespace_name'], k8s_pod_name=$kubernetes['pod_name'], k8s_container_name=$kubernetes['container_name']\n"` |  |
| fluent-bit.config.service | string | `"[SERVICE]\n    Daemon Off\n    Flush {{ .Values.flush }}\n    Log_Level {{ .Values.logLevel }}\n    Parsers_File parsers.conf\n    Parsers_File custom_parsers.conf\n    HTTP_Server On\n    HTTP_Listen 0.0.0.0\n    HTTP_Port {{ .Values.metricsPort }}\n    Health_Check On\n"` |  |
| fluent-bit.enabled | bool | `true` |  |
| fluent-bit.fullnameOverride | string | `"fluent-bit"` |  |
| fluent-bit.image.repository | string | `"fluent/fluent-bit"` |  |
| fluent-bit.podAnnotations."fluentbit.io/exclude" | string | `"true"` |  |
| fluent-bit.podAnnotations."prometheus.io/path" | string | `"/api/v1/metrics/prometheus"` |  |
| fluent-bit.podAnnotations."prometheus.io/port" | string | `"2020"` |  |
| fluent-bit.podAnnotations."prometheus.io/scrape" | string | `"true"` |  |
| fluent-bit.testFramework.enabled | bool | `false` |  |
| grafana.adminPassword | string | `"admin"` |  |
| grafana.dashboardProviders."dashboardproviders.yaml".apiVersion | int | `1` |  |
| grafana.dashboardProviders."dashboardproviders.yaml".providers[0].disableDeletion | bool | `false` |  |
| grafana.dashboardProviders."dashboardproviders.yaml".providers[0].editable | bool | `true` |  |
| grafana.dashboardProviders."dashboardproviders.yaml".providers[0].folder | string | `"envoy-gateway"` |  |
| grafana.dashboardProviders."dashboardproviders.yaml".providers[0].name | string | `"envoy-gateway"` |  |
| grafana.dashboardProviders."dashboardproviders.yaml".providers[0].options.path | string | `"/var/lib/grafana/dashboards/envoy-gateway"` |  |
| grafana.dashboardProviders."dashboardproviders.yaml".providers[0].orgId | int | `1` |  |
| grafana.dashboardProviders."dashboardproviders.yaml".providers[0].type | string | `"file"` |  |
| grafana.dashboardsConfigMaps.envoy-gateway | string | `"grafana-dashboards"` |  |
| grafana.datasources."datasources.yaml".apiVersion | int | `1` |  |
| grafana.datasources."datasources.yaml".datasources[0].name | string | `"Prometheus"` |  |
| grafana.datasources."datasources.yaml".datasources[0].type | string | `"prometheus"` |  |
| grafana.datasources."datasources.yaml".datasources[0].url | string | `"http://prometheus"` |  |
| grafana.enabled | bool | `true` |  |
| grafana.fullnameOverride | string | `"grafana"` |  |
| grafana.service.type | string | `"LoadBalancer"` |  |
| grafana.testFramework.enabled | bool | `false` |  |
| loki.backend.replicas | int | `0` |  |
| loki.deploymentMode | string | `"SingleBinary"` |  |
| loki.enabled | bool | `true` |  |
| loki.fullnameOverride | string | `"loki"` |  |
| loki.gateway.enabled | bool | `false` |  |
| loki.loki.auth_enabled | bool | `false` |  |
| loki.loki.commonConfig.replication_factor | int | `1` |  |
| loki.loki.compactorAddress | string | `"loki"` |  |
| loki.loki.memberlist | string | `"loki-memberlist"` |  |
| loki.loki.rulerConfig.storage.type | string | `"local"` |  |
| loki.loki.storage.type | string | `"filesystem"` |  |
| loki.monitoring.lokiCanary.enabled | bool | `false` |  |
| loki.monitoring.selfMonitoring.enabled | bool | `false` |  |
| loki.monitoring.selfMonitoring.grafanaAgent.installOperator | bool | `false` |  |
| loki.read.replicas | int | `0` |  |
| loki.singleBinary.replicas | int | `1` |  |
| loki.test.enabled | bool | `false` |  |
| loki.write.replicas | int | `0` |  |
| opentelemetry-collector.config.exporters.debug.verbosity | string | `"detailed"` |  |
| opentelemetry-collector.config.exporters.loki.endpoint | string | `"http://loki.monitoring.svc:3100/loki/api/v1/push"` |  |
| opentelemetry-collector.config.exporters.otlp.endpoint | string | `"tempo.monitoring.svc:4317"` |  |
| opentelemetry-collector.config.exporters.otlp.tls.insecure | bool | `true` |  |
| opentelemetry-collector.config.exporters.prometheus.endpoint | string | `"[${env:MY_POD_IP}]:19001"` |  |
| opentelemetry-collector.config.extensions.health_check.endpoint | string | `"[${env:MY_POD_IP}]:13133"` |  |
| opentelemetry-collector.config.processors.attributes.actions[0].action | string | `"insert"` |  |
| opentelemetry-collector.config.processors.attributes.actions[0].key | string | `"loki.attribute.labels"` |  |
| opentelemetry-collector.config.processors.attributes.actions[0].value | string | `"k8s.pod.name, k8s.namespace.name"` |  |
| opentelemetry-collector.config.receivers.datadog.endpoint | string | `"[${env:MY_POD_IP}]:8126"` |  |
| opentelemetry-collector.config.receivers.envoyals.endpoint | string | `"[${env:MY_POD_IP}]:9000"` |  |
| opentelemetry-collector.config.receivers.jaeger.protocols.grpc.endpoint | string | `"[${env:MY_POD_IP}]:14250"` |  |
| opentelemetry-collector.config.receivers.jaeger.protocols.thrift_compact.endpoint | string | `"[${env:MY_POD_IP}]:6831"` |  |
| opentelemetry-collector.config.receivers.jaeger.protocols.thrift_http.endpoint | string | `"[${env:MY_POD_IP}]:14268"` |  |
| opentelemetry-collector.config.receivers.otlp.protocols.grpc.endpoint | string | `"[${env:MY_POD_IP}]:4317"` |  |
| opentelemetry-collector.config.receivers.otlp.protocols.http.endpoint | string | `"[${env:MY_POD_IP}]:4318"` |  |
| opentelemetry-collector.config.receivers.prometheus.config.scrape_configs[0].job_name | string | `"opentelemetry-collector"` |  |
| opentelemetry-collector.config.receivers.prometheus.config.scrape_configs[0].scrape_interval | string | `"10s"` |  |
| opentelemetry-collector.config.receivers.prometheus.config.scrape_configs[0].static_configs[0].targets[0] | string | `"[${env:MY_POD_IP}]:8888"` |  |
| opentelemetry-collector.config.receivers.zipkin.endpoint | string | `"[${env:MY_POD_IP}]:9411"` |  |
| opentelemetry-collector.config.service.extensions[0] | string | `"health_check"` |  |
| opentelemetry-collector.config.service.pipelines.logs.exporters[0] | string | `"loki"` |  |
| opentelemetry-collector.config.service.pipelines.logs.processors[0] | string | `"attributes"` |  |
| opentelemetry-collector.config.service.pipelines.logs.receivers[0] | string | `"otlp"` |  |
| opentelemetry-collector.config.service.pipelines.logs.receivers[1] | string | `"envoyals"` |  |
| opentelemetry-collector.config.service.pipelines.metrics.exporters[0] | string | `"prometheus"` |  |
| opentelemetry-collector.config.service.pipelines.metrics.receivers[0] | string | `"datadog"` |  |
| opentelemetry-collector.config.service.pipelines.metrics.receivers[1] | string | `"otlp"` |  |
| opentelemetry-collector.config.service.pipelines.traces.exporters[0] | string | `"otlp"` |  |
| opentelemetry-collector.config.service.pipelines.traces.receivers[0] | string | `"datadog"` |  |
| opentelemetry-collector.config.service.pipelines.traces.receivers[1] | string | `"otlp"` |  |
| opentelemetry-collector.config.service.pipelines.traces.receivers[2] | string | `"zipkin"` |  |
| opentelemetry-collector.config.service.telemetry.metrics.address | string | `"[${env:MY_POD_IP}]:8888"` |  |
| opentelemetry-collector.enabled | bool | `false` |  |
| opentelemetry-collector.fullnameOverride | string | `"otel-collector"` |  |
| opentelemetry-collector.image.repository | string | `"otel/opentelemetry-collector-contrib"` |  |
| opentelemetry-collector.image.tag | string | `"0.121.0"` |  |
| opentelemetry-collector.mode | string | `"deployment"` |  |
| opentelemetry-collector.ports.envoy-als.appProtocol | string | `"grpc"` |  |
| opentelemetry-collector.ports.envoy-als.containerPort | int | `9000` |  |
| opentelemetry-collector.ports.envoy-als.enabled | bool | `true` |  |
| opentelemetry-collector.ports.envoy-als.hostPort | int | `9000` |  |
| opentelemetry-collector.ports.envoy-als.protocol | string | `"TCP"` |  |
| opentelemetry-collector.ports.envoy-als.servicePort | int | `9000` |  |
| prometheus.alertmanager.enabled | bool | `false` |  |
| prometheus.enabled | bool | `true` |  |
| prometheus.kube-state-metrics.enabled | bool | `false` |  |
| prometheus.prometheus-node-exporter.enabled | bool | `false` |  |
| prometheus.prometheus-pushgateway.enabled | bool | `false` |  |
| prometheus.server.fullnameOverride | string | `"prometheus"` |  |
| prometheus.server.global.scrape_interval | string | `"15s"` |  |
| prometheus.server.image.repository | string | `"prom/prometheus"` |  |
| prometheus.server.persistentVolume.enabled | bool | `false` |  |
| prometheus.server.readinessProbeInitialDelay | int | `0` |  |
| prometheus.server.securityContext | object | `{}` |  |
| prometheus.server.service.type | string | `"LoadBalancer"` |  |
| tempo.enabled | bool | `true` |  |
| tempo.fullnameOverride | string | `"tempo"` |  |
| tempo.service.type | string | `"LoadBalancer"` |  |

